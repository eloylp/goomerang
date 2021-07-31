package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/engine"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/message/protocol"
)

type Handler func(ops Ops, msg proto.Message) error

type Server struct {
	intServer    *http.Server
	c            *websocket.Conn
	upgrader     *websocket.Upgrader
	serverOpts   *serverOpts
	registry     engine.Registry
	errorHandler func(err error)
}

func NewServer(opts ...Option) (*Server, error) {
	cfg := &Config{
		ErrorHandler: func(err error) {},
	}
	for _, o := range opts {
		o(cfg)
	}
	s := &Server{
		upgrader: &websocket.Upgrader{},
		intServer: &http.Server{
			Addr: cfg.ListenURL,
		},
		errorHandler: cfg.ErrorHandler,
		registry:     engine.Registry{},
	}
	s.serverOpts = &serverOpts{s}
	return s, nil
}

func (s *Server) ServerMainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		s.c = c
		defer c.Close()
		for {
			m, data, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			if m == websocket.BinaryMessage {
				frame := &protocol.Frame{}
				if err = proto.Unmarshal(data, frame); err != nil {
					log.Println("parsing: ", err)
				}
				msg, handlers, err := s.registry.Handler(frame.GetType())
				if err != nil {
					log.Println("server handler err: ", err)
				}
				for _, h := range handlers {
					if err = h.(Handler)(s.serverOpts, msg); err != nil {
						s.errorHandler(fmt.Errorf("server handler err: %w", err))
					}
				}
			}
			if m == websocket.CloseMessage {
				s.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				break
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
	marshal, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	envelope := &protocol.Frame{
		Type:    message.FQDN(msg),
		Payload: marshal,
	}
	bytes, err := proto.Marshal(envelope)
	if err != nil {
		return err
	}
	return s.c.WriteMessage(websocket.BinaryMessage, bytes)
}

func (s *Server) Run() error {
	s.intServer.Handler = s.ServerMainHandler()
	return s.intServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.intServer.Shutdown(ctx)
}
