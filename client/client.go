package client

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/engine"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/message/protocol"
)

type Handler func(ops Ops, msg proto.Message) error

type Client struct {
	ServerURL      url.URL
	registry       engine.Registry
	clientOps      *clientOps
	c              *websocket.Conn
	dialer         *websocket.Dialer
	onCloseHandler func()
}

func NewClient(opts ...Option) (*Client, error) {
	cfg := &Config{}
	cfg.OnCloseHandler = func() {}
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
		registry: engine.Registry{},
	}
	c.clientOps = &clientOps{c: c}
	return c, nil
}

func (c *Client) Connect(ctx context.Context) error {
	conn, _, err := c.dialer.DialContext(ctx, c.ServerURL.String(), nil)
	if err != nil {
		return err
	}
	c.c = conn

	go c.startReceiver()

	return nil
}

func (c *Client) startReceiver() {
	func() {
		for {
			m, msg, err := c.c.ReadMessage()
			if err != nil {
				// todo, maybe the error is not assertable. Precheck.
				if err.(*websocket.CloseError).Code == websocket.CloseNormalClosure {
					_ = c.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					c.onCloseHandler()
					return
				}
				log.Println("read:", err)
				return
			}
			if m == websocket.BinaryMessage {
				frame := &protocol.Frame{}
				err = proto.Unmarshal(msg, frame)
				if err != nil {
					log.Println("err on client  receiver:", err)
				}
				msg, handlers, err := c.registry.Handler(frame.Type)
				if err != nil {
					log.Println("client handler err: ", err)
				}
				if err := proto.Unmarshal(frame.Payload, msg); err != nil {
					log.Println("decode: ", err)
				}
				for _, h := range handlers {
					if err = h.(Handler)(c.clientOps, msg); err != nil {
						log.Println("client handler err: ", err)
					}
				}
			}
		}
	}()
}

func (c *Client) Send(ctx context.Context, msg proto.Message) error {
	userMessage, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	frame := &protocol.Frame{
		Type:    message.FQDN(msg),
		Payload: userMessage,
	}
	data, err := proto.Marshal(frame)
	if err != nil {
		return err
	}
	return c.c.WriteMessage(websocket.BinaryMessage, data)
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
