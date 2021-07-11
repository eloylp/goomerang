package goomerang

import (
	"context"
	"github.com/gorilla/websocket"
	messages2 "go.eloylp.dev/goomerang/messages"
	"google.golang.org/protobuf/proto"
	"log"
	"net/url"
)

type Client struct {
	ServerURL url.URL
	c         *websocket.Conn
	handler   ClientHandler
}

type ClientHandler func(client *Client, msg proto.Message) error

func NewClient(opts ...ClientOption) (*Client, error) {
	cfg := &ClientConfig{}
	for _, o := range opts {
		o(cfg)
	}
	serverURL := url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"}
	return &Client{
		ServerURL: serverURL,
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.Dial(c.ServerURL.String(), nil)
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
			m, message, err := c.c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if m == websocket.BinaryMessage {
				envelope := &messages2.Message{}
				err = proto.Unmarshal(message, envelope)
				if err != nil {
					log.Println("err on client  receiver:", err)
				}

				switch envelope.Type {
				case "goomerang.messages.PingPong":
					pingpongMessage := &messages2.PingPong{}
					err := proto.Unmarshal(envelope.Payload, pingpongMessage)
					if err != nil {
						log.Println("err on client  receiver:", err)
					}
					err = c.handler(c, pingpongMessage)
					if err != nil {
						log.Println("err on client  receiver:", err)
					}
				}
			}
		}
	}()
}

func (c *Client) Send(ctx context.Context, msg *messages2.PingPong) error {
	userMessage, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	message := &messages2.Message{
		Type:    string(msg.ProtoReflect().Descriptor().FullName()),
		Payload: userMessage,
	}
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return c.c.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) Close() error {
	return c.c.Close()
}

func (c *Client) RegisterHandler(msg *messages2.PingPong, handler ClientHandler) {
	c.handler = handler
}
