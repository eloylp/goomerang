package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/engine"
	"go.eloylp.dev/goomerang/internal/message"
	"go.eloylp.dev/goomerang/internal/message/protocol"
)

type Handler func(ops Ops, msg proto.Message) *HandlerError

type Server struct {
	intServer       *http.Server
	c               map[*websocket.Conn]struct{}
	L               *sync.Mutex
	wg              *sync.WaitGroup
	ctx             context.Context
	cancl           context.CancelFunc
	upgrader        *websocket.Upgrader
	handlerRegistry engine.AppendableRegistry
	messageRegistry message.Registry
	onErrorHandler  func(err error)
	onCloseHandler  func()
}

func NewServer(opts ...Option) (*Server, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	ctx, cancl := context.WithCancel(context.Background())
	s := &Server{
		upgrader: &websocket.Upgrader{},
		intServer: &http.Server{
			Addr: cfg.ListenURL,
		},
		onErrorHandler:  cfg.ErrorHandler,
		onCloseHandler:  cfg.OnCloseHandler,
		handlerRegistry: engine.AppendableRegistry{},
		messageRegistry: message.Registry{},
		c:               map[*websocket.Conn]struct{}{},
		L:               &sync.Mutex{},
		cancl:           cancl,
		wg:              &sync.WaitGroup{},
		ctx:             ctx,
	}
	return s, nil
}

func (s *Server) ServerMainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.wg.Add(1)
		defer s.wg.Done()
		c, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.onErrorHandler(err)
			return
		}
		// Todo, extract this its own function "register connection"
		s.L.Lock()
		s.c[c] = struct{}{}
		s.L.Unlock()
		sOpts := &serverOpts{s, c}
		defer s.closeConnection(c)

		messageReader := s.readMessages(c)

		for {
			select {
			case <-s.ctx.Done():
				s.onErrorHandler(s.ctx.Err())
				return
			case msg, ok := <-messageReader:
				if !ok {
					return
				}
				if msg.mType == websocket.BinaryMessage {
					if err := s.processMessage(c, msg.data, sOpts); err != nil {
						s.onErrorHandler(err)
						continue
					}
				}
			}
		}
	}
}

func (s *Server) readMessages(c *websocket.Conn) chan *receivedMessage {
	ch := make(chan *receivedMessage)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			messageType, data, err := c.ReadMessage()
			if err != nil {
				var closeErr *websocket.CloseError
				if errors.As(err, &closeErr) {
					if closeErr.Code == websocket.CloseNormalClosure {
						_ = s.sendClosingSignal(c)
					}
				} else {
					s.onErrorHandler(err)
				}
				close(ch)
				return
			}
			ch <- &receivedMessage{
				mType: messageType,
				data:  data,
			}
		}
	}()
	return ch
}

type receivedMessage struct {
	mType int
	data  []byte
}

func (s *Server) processMessage(c *websocket.Conn, data []byte, sOpts Ops) error {
	frame, err := message.UnPack(data)
	if err != nil {
		return err
	}
	msg, err := s.messageRegistry.Message(frame.Type)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(frame.Payload, msg); err != nil {
		return err
	}
	handlers, err := s.handlerRegistry.Elems(frame.Type)
	if err != nil {
		return err
	}
	if frame.IsRpc {
		if err := s.doRPC(handlers, c, frame.Uuid, msg); err != nil {
			return err
		}
		return nil
	}
	s.processAsync(handlers, sOpts, msg)
	return nil
}

func (s *Server) doRPC(handlers []interface{}, c *websocket.Conn, frameUUID string, msg proto.Message) error {
	ops := &serverOptsReqRep{}
	multiReply := &protocol.MultiReply{}
	replies := make([]*protocol.MultiReply_Reply, len(handlers))
	for i := 0; i < len(handlers); i++ {
		replies[i] = &protocol.MultiReply_Reply{}
		if err := handlers[i].(Handler)(ops, msg); err != nil {
			replies[i].Error = &protocol.MultiReply_Error{
				Message: err.Message,
				Code:    err.Code,
			}
			continue
		}
		data, err := proto.Marshal(ops.Index(i))
		if err != nil {
			return err
		}
		replies[i].Message = data
		replies[i].MessageType = message.FQDN(msg)
	}
	multiReply.Replies = replies
	responseMsg, err := message.Pack(multiReply, message.FrameWithUUID(frameUUID), message.FrameIsRPC())
	if err != nil {
		return err
	}
	if err := s.writeMessage(c, responseMsg); err != nil {
		return err
	}
	return nil
}

func (s *Server) writeMessage(c *websocket.Conn, responseMsg []byte) error {
	return c.WriteMessage(websocket.BinaryMessage, responseMsg)
}

func (s *Server) processAsync(handlers []interface{}, ops Ops, msg proto.Message) {
	for _, h := range handlers {
		if err := h.(Handler)(ops, msg); err != nil {
			s.onErrorHandler(fmt.Errorf("server handler err: %w", err))
		}
	}
}

func (s *Server) RegisterHandler(msg proto.Message, handlers ...Handler) {
	his := make([]interface{}, len(handlers))
	for i := 0; i < len(handlers); i++ {
		his[i] = handlers[i]
	}
	fqdn := message.FQDN(msg)
	s.messageRegistry.Register(fqdn, msg)
	s.handlerRegistry.Register(fqdn, his...)
}

func (s *Server) Send(ctx context.Context, msg proto.Message) error {
	bytes, err := message.Pack(msg)
	if err != nil {
		return err
	}
	s.L.Lock()
	defer s.L.Unlock()
	var errList error
	var count int
	for conn := range s.c {
		if err := conn.WriteMessage(websocket.BinaryMessage, bytes); err != nil && count < 100 {
			if errors.Is(err, websocket.ErrCloseSent) {
				err = ErrClientDisconnected
			}
			errList = multierror.Append(err, errList)
			count++
		}
	}
	if errList != nil {
		return errList
	}
	return nil
}

func (s *Server) Run() error {
	s.intServer.Handler = s.ServerMainHandler()
	if err := s.intServer.ListenAndServe(); err != http.ErrServerClosed {
		s.onErrorHandler(err)
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.L.Lock()
	defer s.L.Unlock()
	var multiErr error
	if err := s.intServer.Shutdown(ctx); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}
	s.cancl()
	ch := make(chan struct{})
	go func() {
		s.wg.Wait()
		ch <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		multiErr = multierror.Append(multiErr, ctx.Err())
	case <-ch:
	}
	s.onCloseHandler()
	return multiErr
}

func (s *Server) closeConnection(c *websocket.Conn) {
	delete(s.c, c) // todo move this to upper level and extract to its function
	if err := s.sendClosingSignal(c); err != nil {
		s.onErrorHandler(err)
	}
}

func (s *Server) sendClosingSignal(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
