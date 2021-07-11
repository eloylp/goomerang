package goomerang

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/message"
)

var serverUpgrader = websocket.Upgrader{}

func NewServer(opts ...ServerOption) (*Server, error) {
	cfg := &ServerConfig{}
	for _, o := range opts {
		o(cfg)
	}
	s := &Server{
		intServer: &http.Server{
			Addr: cfg.ListenURL,
		},
	}
	return s, nil
}

func (s *Server) ServerMainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := serverUpgrader.Upgrade(w, r, nil)
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
				err := s.handler(s, pingpongMessage)
				if err != nil {
					log.Println("server handler err: ", err)
				}
			}
		}
	}
}

type Server struct {
	intServer *http.Server
	c         *websocket.Conn
	handler   ServerHandler
}

type ServerHandler func(server *Server, msg proto.Message) error

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
