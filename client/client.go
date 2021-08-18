package client

import (
	"context"
	"errors"
	"go.eloylp.dev/goomerang/internal/handler"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
)

type Handler func(ops Ops, msg proto.Message) error

type Client struct {
	ServerURL       url.URL
	registry        handler.Registry
	messageRegistry message.MessageRegistry
	clientOps       *clientOps
	c               *websocket.Conn
	dialer          *websocket.Dialer
	onCloseHandler  func()
	onErrorHandler  func(err error)
	reqRepRegistry  map[string]chan proto.Message
	ids             uint64
}

func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	serverURL := url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"}
	c := &Client{
		ServerURL:      serverURL,
		onCloseHandler: cfg.OnCloseHandler,
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second, // TODO parametrize this.
		},
		registry:        handler.Registry{},
		messageRegistry: message.MessageRegistry{},
		reqRepRegistry:  map[string]chan proto.Message{},
	}
	c.clientOps = &clientOps{c: c}
	return c, nil
}

func (c *Client) Connect(ctx context.Context) error {
	conn, resp, err := c.dialer.DialContext(ctx, c.ServerURL.String(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.c = conn

	go c.startReceiver()

	return nil
}

func (c *Client) startReceiver() {
	func() {
		for {
			m, data, err := c.c.ReadMessage()
			if err != nil {
				var closeErr *websocket.CloseError
				if errors.As(err, &closeErr) {
					if closeErr.Code == websocket.CloseNormalClosure {
						_ = c.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						c.onCloseHandler()
						return
					}
				}
				c.onErrorHandler(err)
				return
			}
			if m == websocket.BinaryMessage {
				frame, err := message.UnPack(data)
				if err != nil {
					c.onErrorHandler(err)
					continue
				}
				msg, err := c.messageRegistry.Message(frame.Type)
				if err != nil {
					c.onErrorHandler(err)
					continue
				}
				if err := proto.Unmarshal(frame.Payload, msg); err != nil {
					c.onErrorHandler(err)
					continue
				}
				if frame.Uuid != "" {
					ch, ok := c.reqRepRegistry[frame.Uuid]
					if !ok {
						c.onErrorHandler(errors.New("frame is marked for req/rep tracking, but no channel receiver found in registry"))
						continue
					}
					ch <- msg
					delete(c.reqRepRegistry, frame.Uuid)
					continue
				}
				handlers, err := c.registry.Handler(frame.Type)
				if err != nil {
					c.onErrorHandler(err)
					continue
				}
				for _, h := range handlers {
					if err = h.(Handler)(c.clientOps, msg); err != nil {
						c.onErrorHandler(err)
					}
				}
			}
		}
	}()
}

func (c *Client) Send(ctx context.Context, msg proto.Message) error {
	data, err := message.Pack(msg)
	if err != nil {
		return err
	}
	if err := c.c.WriteMessage(websocket.BinaryMessage, data); err != nil {
		if errors.Is(err, websocket.ErrCloseSent) {
			return ErrServerDisconnected
		}
		return err
	}
	return nil
}

func (c *Client) Close() error {
	err := c.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		if errors.Is(err, websocket.ErrCloseSent) {
			return ErrServerDisconnected
		}
		return err
	}
	return nil
}

func (c *Client) RegisterHandler(msg proto.Message, handlers ...Handler) {
	his := make([]interface{}, len(handlers))
	for i := 0; i < len(handlers); i++ {
		his[i] = handlers[i]
	}
	fqdn := message.FQDN(msg)
	c.messageRegistry.Register(fqdn, msg)
	c.registry.Register(fqdn, his...)
}

func (c *Client) SendSync(ctx context.Context, msg *testMessages.PingPong) (proto.Message, error) {
	msgUuid := uuid.New().String()
	data, err := message.Pack(msg, message.FrameWithUuid(msgUuid))
	if err != nil {
		return nil, err
	}
	resCh, err := c.interceptFrameID(msgUuid)
	if err != nil {
		return nil, err
	}
	if err := c.c.WriteMessage(websocket.BinaryMessage, data); err != nil {
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil, ErrServerDisconnected
		}
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case respMsg := <-resCh:
		return respMsg, nil
	}
}

func (c *Client) interceptFrameID(uuid string) (chan proto.Message, error) {
	repCh := make(chan proto.Message, 1)
	c.reqRepRegistry[uuid] = repCh
	return repCh, nil
}

func (c *Client) RegisterMessage(msg proto.Message) {
	c.messageRegistry.Register(message.FQDN(msg), msg)
}
