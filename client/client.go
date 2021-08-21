package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/engine"
	"go.eloylp.dev/goomerang/internal/message"
	"go.eloylp.dev/goomerang/internal/message/protocol"
	"go.eloylp.dev/goomerang/server"
)

type Handler func(ops Ops, msg proto.Message) error

type Client struct {
	ServerURL       url.URL
	handlerRegistry engine.AppendableRegistry
	messageRegistry message.Registry
	clientOps       *clientOps
	conn            *websocket.Conn
	dialer          *websocket.Dialer
	onCloseHandler  func()
	onErrorHandler  func(err error)
	reqRepRegistry  map[string]chan *MultiReply
}

func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	c := &Client{
		ServerURL:      url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"},
		onCloseHandler: cfg.OnCloseHandler,
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second, // TODO parametrize this.
		},
		handlerRegistry: engine.AppendableRegistry{},
		messageRegistry: message.Registry{},
		reqRepRegistry:  map[string]chan *MultiReply{},
	}
	c.clientOps = &clientOps{c: c}
	c.RegisterMessage(&protocol.MultiReply{})
	return c, nil
}

func (c *Client) Connect(ctx context.Context) error {
	conn, resp, err := c.dialer.DialContext(ctx, c.ServerURL.String(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.conn = conn

	go c.startReceiver()

	return nil
}

func (c *Client) startReceiver() {
	func() {
		for {
			m, data, err := c.conn.ReadMessage()
			if err != nil {
				var closeErr *websocket.CloseError
				if errors.As(err, &closeErr) {
					if closeErr.Code == websocket.CloseNormalClosure {
						_ = c.sendClosingSignal()
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
				if frame.IsRpc {
					if err := c.doRPC(frame.Uuid, msg); err != nil {
						c.onErrorHandler(err)
					}
					continue
				}
				handlers, err := c.handlerRegistry.Elems(frame.Type)
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

func (c *Client) sendClosingSignal() error {
	return c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (c *Client) doRPC(frameUUID string, msg proto.Message) error {
	ch, ok := c.reqRepRegistry[frameUUID]
	if !ok {
		return errors.New("frame is marked for req/rep tracking, but no channel receiver found in registry")
	}
	protoMultiReply, ok := msg.(*protocol.MultiReply)
	if !ok {
		return errors.New("frame is marked for req/rep tracking, cannot cast to multi-reply message")
	}
	repliesCount := len(protoMultiReply.Replies)
	multiReply := &MultiReply{}
	replies := make([]*Reply, repliesCount)
	multiReply.Replies = replies
	for i := 0; i < repliesCount; i++ {
		replies[i] = &Reply{}
		if protoMultiReply.Replies[i].Error != nil {
			replies[i].Err = server.NewHandlerErrorWith(
				protoMultiReply.Replies[i].Error.Message,
				protoMultiReply.Replies[i].Error.Code,
			)
			continue
		}
		protoMsg, err := c.messageRegistry.Message(protoMultiReply.Replies[i].MessageType)
		if err != nil {
			return errors.New("error parsing message in multi-reply")
		}
		if err := proto.Unmarshal(protoMultiReply.Replies[i].Message, protoMsg); err != nil {
			return errors.New("error parsing message in multi-reply")
		}
		replies[i].Message = protoMsg
	}
	ch <- multiReply
	delete(c.reqRepRegistry, frameUUID)
	return nil
}

func (c *Client) Send(ctx context.Context, msg proto.Message) error {
	data, err := message.Pack(msg)
	if err != nil {
		return err
	}
	if err := c.writeMessage(data); err != nil {
		if errors.Is(err, websocket.ErrCloseSent) {
			return ErrServerDisconnected
		}
		return err
	}
	return nil
}

func (c *Client) writeMessage(data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) Close() error {
	err := c.sendClosingSignal()
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
	c.handlerRegistry.Register(fqdn, his...)
}

func (c *Client) RPC(ctx context.Context, msg proto.Message) (*MultiReply, error) {
	msgUUID := uuid.New().String()
	data, err := message.Pack(msg, message.FrameWithUUID(msgUUID), message.FrameIsRPC())
	if err != nil {
		return nil, err
	}
	resCh := c.interceptFrameID(msgUUID)
	if err := c.writeMessage(data); err != nil {
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil, ErrServerDisconnected
		}
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case multiReply := <-resCh:
		return multiReply, nil
	}
}

func (c *Client) interceptFrameID(uid string) chan *MultiReply {
	repCh := make(chan *MultiReply, 1)
	c.reqRepRegistry[uid] = repCh
	return repCh
}

func (c *Client) RegisterMessage(msg proto.Message) {
	c.messageRegistry.Register(message.FQDN(msg), msg)
}
