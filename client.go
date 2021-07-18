package goomerang

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/message"
)

type Client struct {
	ServerURL url.URL
	handler   ClientHandler
	clientOps *clientOps
	c         *websocket.Conn
	dialer    *websocket.Dialer
}

type ClientOps interface {
	PeerOps
}

type clientOps struct {
	c *Client
}

func (co *clientOps) Send(ctx context.Context, msg proto.Message) error {
	return co.c.Send(ctx, msg)
}

type ClientHandler func(clientOps ClientOps, msg proto.Message) error

func NewClient(opts ...ClientOption) (*Client, error) {
	cfg := &ClientConfig{}
	for _, o := range opts {
		o(cfg)
	}
	serverURL := url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"}
	c := &Client{
		ServerURL: serverURL,
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second, // TODO parametrize this.
		},
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
				log.Println("read:", err)
				return
			}
			if m == websocket.BinaryMessage {
				frame := &message.Frame{}
				err = proto.Unmarshal(msg, frame)
				if err != nil {
					log.Println("err on client  receiver:", err)
				}
				switch frame.Type {
				case "goomerang.test.PingPong":
					pingpongMessage := &message.PingPong{}
					err := proto.Unmarshal(frame.Payload, pingpongMessage)
					if err != nil {
						log.Println("err on client  receiver:", err)
					}
					err = c.handler(c.clientOps, pingpongMessage)
					if err != nil {
						log.Println("err on client  receiver:", err)
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
	frame := &message.Frame{
		Type:    string(msg.ProtoReflect().Descriptor().FullName()),
		Payload: userMessage,
	}
	data, err := proto.Marshal(frame)
	if err != nil {
		return err
	}
	return c.c.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) Close() error {
	return c.c.Close()
}

func (c *Client) RegisterHandler(msg proto.Message, handler ClientHandler) {
	c.handler = handler
}
