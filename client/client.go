package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/conc"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
)

type Client struct {
	ServerURL       url.URL
	handlerChainer  *messaging.HandlerChainer
	messageRegistry messaging.Registry
	writeLock       *sync.Mutex
	conn            *websocket.Conn
	dialer          *websocket.Dialer
	onCloseHook     func()
	onErrorHook     func(err error)
	requestRegistry *requestRegistry
	workerPool      *conc.WorkerPool
	closeCh         chan struct{}
}

func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	wp, err := conc.NewWorkerPool(cfg.MaxConcurrency)
	if err != nil {
		return nil, fmt.Errorf("goomerang client: %w", err)
	}
	c := &Client{
		ServerURL:   serverURL(cfg),
		onCloseHook: cfg.OnCloseHook,
		onErrorHook: cfg.OnErrorHook,
		dialer: &websocket.Dialer{
			Proxy:             http.ProxyFromEnvironment,
			TLSClientConfig:   cfg.TLSConfig,
			HandshakeTimeout:  cfg.HandshakeTimeout,
			ReadBufferSize:    cfg.ReadBufferSize,
			WriteBufferSize:   cfg.WriteBufferSize,
			EnableCompression: cfg.EnableCompression,
		},
		writeLock:       &sync.Mutex{},
		handlerChainer:  messaging.NewHandlerChainer(),
		messageRegistry: messaging.Registry{},
		requestRegistry: newRegistry(),
		workerPool:      wp,
		closeCh:         make(chan struct{}),
	}
	return c, nil
}

func (c *Client) Connect(ctx context.Context) error {
	c.handlerChainer.PrepareChains()
	conn, resp, err := c.dialer.DialContext(ctx, c.ServerURL.String(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.conn = conn
	go c.receiver()
	return nil
}

func (c *Client) receiver() {
	for {
		select {
		case <-c.closeCh:
			return
		default:
			messageType, data, err := c.conn.ReadMessage()
			if err != nil {
				var closeErr *websocket.CloseError
				if errors.As(err, &closeErr) {
					if closeErr.Code == websocket.CloseNormalClosure {
						c.onErrorHook(c.Close(context.Background()))
						return
					}
				}
				c.onErrorHook(err)
				return
			}
			if messageType != websocket.BinaryMessage {
				c.onErrorHook(fmt.Errorf("protocol: unexpected message type %v", messageType))
				return
			}
			c.workerPool.Add()
			go func() {
				defer c.workerPool.Done()
				if err := c.processMessage(data); err != nil {
					c.onErrorHook(err)
				}
			}()
		}
	}
}

func (c *Client) processMessage(data []byte) (err error) {
	frame, err := messaging.UnPack(data)
	if err != nil {
		return err
	}
	msg, err := messaging.FromFrame(frame, c.messageRegistry)
	if err != nil {
		return err
	}
	if msg.Metadata.IsSync {
		if err := c.receiveSync(msg); err != nil {
			return err
		}
		return nil
	}
	handler, err := c.handlerChainer.Handler(msg.Metadata.Type)
	if err != nil {
		return err
	}
	handler.Handle(c, msg)
	return nil
}

func (c *Client) Send(msg *message.Message) (payloadSize int, err error) {
	payloadSize, data, err := messaging.Pack(msg)
	if err != nil {
		return payloadSize, ErrServerDisconnected
	}
	if err := c.writeMessage(data); err != nil {
		if errors.Is(err, websocket.ErrCloseSent) {
			return payloadSize, ErrServerDisconnected
		} else {
			return payloadSize, err
		}
	}
	return
}

func (c *Client) SendSync(ctx context.Context, msg *message.Message) (payloadSize int, response *message.Message, err error) {
	ch := make(chan sendSyncResponse, 1)
	go func() {
		UUID := uuid.New().String()
		payloadSize, data, err := messaging.Pack(msg, messaging.FrameWithUUID(UUID), messaging.FrameIsSync())
		if err != nil {
			ch <- sendSyncResponse{payloadSize, nil, err}
			return
		}
		c.requestRegistry.createListener(UUID)
		if err := c.writeMessage(data); err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				ch <- sendSyncResponse{payloadSize, nil, ErrServerDisconnected}
			}
			ch <- sendSyncResponse{payloadSize, nil, err}
			return
		}
		repliedMsg, err := c.requestRegistry.resultFor(ctx, UUID)
		if err != nil {
			ch <- sendSyncResponse{payloadSize, nil, err}
			return
		}
		ch <- sendSyncResponse{payloadSize, repliedMsg, nil}
	}()
	select {
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	case resp := <-ch:
		if resp.err != nil {
			return 0, nil, resp.err
		}
		return resp.payloadSize, resp.respMsg, resp.err
	}
}

type sendSyncResponse struct {
	payloadSize int
	respMsg     *message.Message
	err         error
}

func (c *Client) Close(ctx context.Context) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		defer c.workerPool.Wait()
		defer c.onCloseHook()
		if err := c.sendClosingSignal(); err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				ch <- ErrServerDisconnected
			} else {
				ch <- err
			}
			return
		}
		if err := c.conn.Close(); err != nil {
			ch <- err
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

func (c *Client) RegisterMiddleware(m message.Middleware) {
	c.handlerChainer.AppendMiddleware(m)
}

func (c *Client) RegisterHandler(msg proto.Message, h message.Handler) {
	fqdn := messaging.FQDN(msg)
	c.messageRegistry.Register(fqdn, msg)
	c.handlerChainer.AppendHandler(fqdn, h)
}

func (c *Client) RegisterMessage(msg proto.Message) {
	c.messageRegistry.Register(messaging.FQDN(msg), msg)
}

func (c *Client) writeMessage(data []byte) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) sendClosingSignal() error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	return c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (c *Client) receiveSync(msg *message.Message) error {
	if err := c.requestRegistry.submitResult(msg.Metadata.UUID, msg); err != nil {
		return err
	}
	return nil
}
