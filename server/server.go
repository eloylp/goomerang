package server

import (
	"context"
	"errors"
	"fmt"
	"go.eloylp.dev/goomerang/internal/message"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type Handler func(ops Ops, msg proto.Message) error

type Server struct {
	intServer      *http.Server
	c              []*websocket.Conn
	L              *sync.Mutex
	upgrader       *websocket.Upgrader
	registry       message.Registry
	errorHandler   func(err error)
	onCloseHandler func()
}

func NewServer(opts ...Option) (*Server, error) {
	cfg := &Config{
		ErrorHandler:   func(err error) {},
		OnCloseHandler: func() {},
	}
	for _, o := range opts {
		o(cfg)
	}
	s := &Server{
		upgrader: &websocket.Upgrader{},
		intServer: &http.Server{
			Addr: cfg.ListenURL,
		},
		errorHandler:   cfg.ErrorHandler,
		onCloseHandler: cfg.OnCloseHandler,
		registry:       message.Registry{},
		L:              &sync.Mutex{},
	}
	return s, nil
}

func (s *Server) ServerMainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		s.L.Lock()
		s.c = append(s.c, c)
		s.L.Unlock()
		sOpts := &serverOpts{s, c}
		defer c.Close()

		for {
			m, data, err := c.ReadMessage()
			if err != nil {
				// todo, maybe the error is not assertable. Precheck.
				if err.(*websocket.CloseError).Code == websocket.CloseNormalClosure {
					_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					s.onCloseHandler()
					return
				}
				log.Println("read:", err)
				break
			}
			if m == websocket.BinaryMessage {
				msg, handlers, err := message.UnPack(s.registry, data)
				if err != nil {
					s.errorHandler(err)
					continue
				}
				for _, h := range handlers {
					if err = h.(Handler)(sOpts, msg); err != nil {
						s.errorHandler(fmt.Errorf("server handler err: %w", err))
					}
				}
			}
		}
	}
}

func (s *Server) RegisterHandler(msg proto.Message, handlers ...Handler) {
	his := make([]interface{}, len(handlers))
	for i, h := range handlers {
		his[i] = h
	}
	s.registry.Register(msg, his...)
}

func (s *Server) Send(ctx context.Context, msg proto.Message) error {
	bytes, err := message.Pack(msg)
	if err != nil {
		return err
	}
	s.L.Lock()
	defer s.L.Unlock()
	var errs strings.Builder
	for _, conn := range s.c {
		if err := conn.WriteMessage(websocket.BinaryMessage, bytes); err != nil {
			errs.WriteString(err.Error() + " \n")
		}
	}
	if errs.Len() != 0 {
		return errors.New("send: \n" + errs.String())
	}
	return nil
}

func (s *Server) Run() error {
	s.intServer.Handler = s.ServerMainHandler()
	return s.intServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// todo this must be done for ALL connections.
	s.c[0].WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// TODO. This must be done gracefully.
	return s.intServer.Shutdown(ctx)
}
