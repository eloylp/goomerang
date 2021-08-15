package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message"
)

type Handler func(ops Ops, msg proto.Message) error

type Client struct {
	ServerURL      url.URL
	registry       message.Registry
	clientOps      *clientOps
	c              *websocket.Conn
	dialer         *websocket.Dialer
	onCloseHandler func()
	onErrorHandler func(err error)
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
		registry: message.Registry{},
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
				msg, handlers, err := message.UnPack(c.registry, data)
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
	return c.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (c *Client) RegisterHandler(msg proto.Message, handlers ...Handler) {
	his := make([]interface{}, len(handlers))
	for i, h := range handlers {
		his[i] = h
	}
	c.registry.Register(msg, his...)
}
