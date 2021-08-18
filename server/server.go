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
)

type Handler func(ops Ops, msg proto.Message) error

type Server struct {
	intServer       *http.Server
	c               map[*websocket.Conn]struct{}
	L               *sync.Mutex
	upgrader        *websocket.Upgrader
	registry        engine.AppendableRegistry
	messageRegistry message.Registry
	onErrorHandler  func(err error)
	onCloseHandler  func()
}

func NewServer(opts ...Option) (*Server, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	s := &Server{
		upgrader: &websocket.Upgrader{},
		intServer: &http.Server{
			Addr: cfg.ListenURL,
		},
		onErrorHandler:  cfg.ErrorHandler,
		onCloseHandler:  cfg.OnCloseHandler,
		registry:        engine.AppendableRegistry{},
		messageRegistry: message.Registry{},
		c:               map[*websocket.Conn]struct{}{},
		L:               &sync.Mutex{},
	}
	return s, nil
}

func (s *Server) ServerMainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.onErrorHandler(err)
			return
		}
		s.L.Lock()
		s.c[c] = struct{}{}
		s.L.Unlock()
		sOpts := &serverOpts{s, c}
		defer c.Close()

		for {
			m, data, err := c.ReadMessage()
			if err != nil {
				var closeErr *websocket.CloseError
				if errors.As(err, &closeErr) {
					if closeErr.Code == websocket.CloseNormalClosure {
						_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						s.onCloseHandler()
						return
					}
				}
				s.onErrorHandler(err)
				break
			}
			if m == websocket.BinaryMessage {
				frame, err := message.UnPack(data)
				if err != nil {
					s.onErrorHandler(err)
					continue
				}
				var currentOpts Ops = sOpts
				if frame.Uuid != "" {
					currentOpts = &serverOptsReqRep{c: c, s: s, frameUuid: frame.Uuid}
				}
				msg, err := s.messageRegistry.Message(frame.Type)
				if err != nil {
					s.onErrorHandler(err)
					continue
				}
				handlers, err := s.registry.Elems(frame.Type)
				if err != nil {
					s.onErrorHandler(err)
					continue
				}
				if err := proto.Unmarshal(frame.Payload, msg); err != nil {
					s.onErrorHandler(err)
					continue
				}
				for _, h := range handlers {
					if err = h.(Handler)(currentOpts, msg); err != nil {
						s.onErrorHandler(fmt.Errorf("server handler err: %w", err))
					}
				}
			}
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
	s.registry.Register(fqdn, his...)
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
	return s.intServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.L.Lock()
	defer s.L.Unlock()
	var errList error
	var count int
	for conn := range s.c {
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil && err != websocket.ErrCloseSent && count < 100 {
			errList = multierror.Append(errList, err)
			count++
		}
	}
	if errList != nil {
		return errList
	}
	// TODO. This must be done gracefully.
	return s.intServer.Shutdown(ctx)
}
