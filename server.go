package goomerang

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/message"
)

func NewServer(opts ...ServerOption) (*Server, error) {
	cfg := &ServerConfig{}
	for _, o := range opts {
		o(cfg)
	}
	s := &Server{
		upgrader: &websocket.Upgrader{},
		intServer: &http.Server{
			Addr: cfg.ListenURL,
		},
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
			_, data, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			frame := &message.Frame{}
			if err = proto.Unmarshal(data, frame); err != nil {
				log.Println("parsing: ", err)
			}
			switch frame.GetType() {
			case "goomerang.test.PingPong":
				pingpongMessage := &message.PingPong{}
				if err := proto.Unmarshal(frame.GetPayload(), pingpongMessage); err != nil {
					log.Println("parsing: ", err)
				}
				err := s.handler(s.serverOpts, pingpongMessage)
				if err != nil {
					log.Println("server handler err: ", err)
				}
			}
		}
	}
}

type Server struct {
	intServer  *http.Server
	c          *websocket.Conn
	upgrader   *websocket.Upgrader
	handler    ServerHandler
	serverOpts *serverOpts
}

type ServerHandler func(serverOpts ServerOpts, msg proto.Message) error

func (s *Server) RegisterHandler(msg proto.Message, handler ServerHandler) {
	s.handler = handler
}

func (s *Server) Send(ctx context.Context, msg proto.Message) error {
	marshal, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	envelope := &message.Frame{
		Type:    string(msg.ProtoReflect().Descriptor().FullName()),
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

type ServerOpts interface {
	PeerOps
	Shutdown(ctx context.Context) error
}

type serverOpts struct {
	s *Server
}

func (so *serverOpts) Send(ctx context.Context, msg proto.Message) error {
	return so.s.Send(ctx, msg)
}

func (so *serverOpts) Shutdown(ctx context.Context) error {
	return so.s.Shutdown(ctx)
}
