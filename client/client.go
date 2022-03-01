package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/client/internal/rpc"
	"go.eloylp.dev/goomerang/internal/message"
)

type Client struct {
	ServerURL       url.URL
	handlerChainer  *message.HandlerChainer
	messageRegistry message.Registry
	clientOps       *immediateSender
	writeLock       *sync.Mutex
	conn            *websocket.Conn
	dialer          *websocket.Dialer
	onCloseHook     func()
	onErrorHook     func(err error)
	rpcRegistry     *rpc.Registry
	workerPool      *goomerang.WorkerPool
	closeCh         chan struct{}
}

func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	wp, err := goomerang.NewWorkerPool(cfg.MaxConcurrency)
	if err != nil {
		return nil, fmt.Errorf("goomerang client: %w", err)
	}
	c := &Client{
		ServerURL:   serverURL(cfg),
		onCloseHook: cfg.OnCloseHook,
		onErrorHook: cfg.OnErrorHook,
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			TLSClientConfig:  cfg.TLSConfig,
			ReadBufferSize:   cfg.ReadBufferSize,
			WriteBufferSize:  cfg.WriteBufferSize,
			HandshakeTimeout: 2 * time.Second,
		},
		writeLock:       &sync.Mutex{},
		handlerChainer:  message.NewHandlerChainer(),
		messageRegistry: message.Registry{},
		rpcRegistry:     rpc.NewRegistry(),
		workerPool:      wp,
		closeCh:         make(chan struct{}),
	}
	c.clientOps = &immediateSender{c: c}
	return c, nil
}

func serverURL(cfg *Config) url.URL {
	if cfg.TLSConfig != nil {
		return url.URL{Scheme: "wss", Host: cfg.TargetServer, Path: "/wss"}
	}
	return url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"}
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
						_ = c.sendClosingSignal()
						c.onCloseHook()
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
				if err := c.processMessage(data); err != nil {
					c.onErrorHook(err)
				}
				c.workerPool.Done()
			}()
		}
	}
}

func (c *Client) processMessage(data []byte) error {
	frame, err := message.UnPack(data)
	if err != nil {
		return err
	}
	msg, err := message.FromFrame(frame, c.messageRegistry)
	if err != nil {
		return err
	}
	if msg.Metadata().IsRPC {
		if err := c.processRPC(msg); err != nil {
			return err
		}
		return nil
	}
	handler, err := c.handlerChainer.Handler(msg.Metadata().Type)
	if err != nil {
		return err
	}
	handler.Handle(c.clientOps, msg)
	return nil
}

func (c *Client) Send(ctx context.Context, msg *message.Message) error {
	data, err := message.Pack(msg)
	if err != nil {
		return err
	}
	ch := make(chan error, 1)
	go func() {
		if err := c.writeMessage(data); err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				ch <- ErrServerDisconnected
			} else {
				ch <- err
			}
		}
		close(ch)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

func (c *Client) Close(ctx context.Context) error {
	ch := make(chan error, 1)
	go func() {
		err := c.sendClosingSignal()
		if err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				ch <- ErrServerDisconnected
			} else {
				ch <- err
			}
		}
		close(c.closeCh)
		c.workerPool.Wait()
		close(ch)
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
	fqdn := message.FQDN(msg)
	c.messageRegistry.Register(fqdn, msg)
	c.handlerChainer.AppendHandler(fqdn, h)
}

func (c *Client) RPC(ctx context.Context, msg *message.Message) (*message.Message, error) {
	UUID := uuid.New().String()
	data, err := message.Pack(msg, message.FrameWithUUID(UUID), message.FrameIsRPC())
	if err != nil {
		return nil, err
	}
	c.rpcRegistry.CreateListener(UUID)
	if err := c.writeMessage(data); err != nil {
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil, ErrServerDisconnected
		}
		return nil, err
	}
	repliedMsg, err := c.rpcRegistry.ResultFor(ctx, UUID)
	if err != nil {
		return nil, err
	}
	return repliedMsg, nil
}

func (c *Client) RegisterMessage(msg proto.Message) {
	c.messageRegistry.Register(message.FQDN(msg), msg)
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

func (c *Client) processRPC(msg *message.Message) error {
	if err := c.rpcRegistry.SubmitResult(msg.Metadata().UUID, msg); err != nil {
		return err
	}
	return nil
}
