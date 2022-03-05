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

	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/internal/conc"
	"go.eloylp.dev/goomerang/internal/message"
)

type connSlot struct {
	l *sync.Mutex
	c *websocket.Conn
}

type Server struct {
	intServer       *http.Server
	connRegistry    map[*websocket.Conn]connSlot
	serverL         *sync.Mutex
	wg              *sync.WaitGroup
	ctx             context.Context
	cancl           context.CancelFunc
	wsUpgrader      *websocket.Upgrader
	handlerChainer  *message.HandlerChainer
	messageRegistry message.Registry
	onErrorHook     func(err error)
	onCloseHook     func()
	cfg             *Config
	workerPool      *conc.WorkerPool
}

func NewServer(opts ...Option) (*Server, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	wp, err := conc.NewWorkerPool(cfg.MaxConcurrency)
	if err != nil {
		return nil, fmt.Errorf("goomerang server: %w", err)
	}
	ctx, cancl := context.WithCancel(context.Background())
	s := &Server{
		wsUpgrader: &websocket.Upgrader{
			HandshakeTimeout:  cfg.HandshakeTimeout,
			ReadBufferSize:    cfg.ReadBufferSize,
			WriteBufferSize:   cfg.WriteBufferSize,
			EnableCompression: cfg.EnableCompression,
		},
		intServer: &http.Server{
			Addr:              cfg.ListenURL,
			TLSConfig:         cfg.TLSConfig,
			ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
			ReadTimeout:       cfg.HTTPReadTimeout,
			WriteTimeout:      cfg.HTTPWriteTimeout,
		},
		onErrorHook:     cfg.OnErrorHook,
		onCloseHook:     cfg.OnCloseHook,
		handlerChainer:  message.NewHandlerChainer(),
		messageRegistry: message.Registry{},
		connRegistry:    map[*websocket.Conn]connSlot{},
		serverL:         &sync.Mutex{},
		cancl:           cancl,
		wg:              &sync.WaitGroup{},
		ctx:             ctx,
		cfg:             cfg,
		workerPool:      wp,
	}
	mux := http.NewServeMux()
	mux.Handle(endpoint(cfg), mainHandler(s))
	s.intServer.Handler = mux
	return s, nil
}

func endpoint(cfg *Config) string {
	if cfg.TLSConfig != nil {
		return "/wss"
	}
	return "/ws"
}

func mainHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.wg.Add(1)
		defer s.wg.Done()
		c, err := s.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			s.onErrorHook(err)
			return
		}
		cs := s.addConnection(c)
		sOpts := &immediateSender{s: s, connSlot: cs}
		defer s.removeConnection(cs)
		defer s.closeConnection(cs)

		messageReader := s.readMessages(cs)

		for {
			select {
			case <-s.ctx.Done():
				return
			case msg, ok := <-messageReader:
				if !ok { // Channel closed in the receiver, we abort this handler and start defer calling.
					return
				}
				if msg.mType != websocket.BinaryMessage {
					s.onErrorHook(fmt.Errorf("server: cannot process websocket frame type %v", msg.mType))
					continue
				}
				s.workerPool.Add() // Will block till more processing slots are available.
				go func() {
					if err := s.processMessage(cs, msg.data, sOpts); err != nil {
						s.onErrorHook(err)
					}
					s.workerPool.Done()
				}()
			}
		}
	}
}

func (s *Server) addConnection(c *websocket.Conn) connSlot {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	slot := connSlot{
		l: &sync.Mutex{},
		c: c,
	}
	s.connRegistry[c] = slot
	return slot
}

func (s *Server) removeConnection(c connSlot) {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	delete(s.connRegistry, c.c)
}

func (s *Server) readMessages(cs connSlot) chan *receivedMessage {
	ch := make(chan *receivedMessage)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				close(ch)
				return
			default:
				messageType, data, err := cs.c.ReadMessage()
				if err != nil {
					var closeErr *websocket.CloseError
					if errors.As(err, &closeErr) {
						if closeErr.Code == websocket.CloseNormalClosure {
							_ = s.sendClosingSignal(cs)
						}
						return
					}
					s.onErrorHook(err)
					close(ch)
					return
				}
				ch <- &receivedMessage{
					mType: messageType,
					data:  data,
				}
			}
		}
	}()
	return ch
}

type receivedMessage struct {
	mType int
	data  []byte
}

func (s *Server) processMessage(cs connSlot, data []byte, sOpts message.Sender) error {
	frame, err := message.UnPack(data)
	if err != nil {
		return err
	}

	msg, err := message.FromFrame(frame, s.messageRegistry)
	if err != nil {
		return err
	}
	handler, err := s.handlerChainer.Handler(msg.Metadata.Type)
	if err != nil {
		return err
	}
	if msg.Metadata.IsRPC {
		if err := s.doRPC(handler, cs, msg); err != nil {
			return err
		}
		return nil
	}
	handler.Handle(sOpts, msg)
	return nil
}

func (s *Server) doRPC(handler message.Handler, cs connSlot, msg *goomerang.Message) error {
	ops := &bufferedSender{}
	handler.Handle(ops, msg)
	responseMsg, err := message.Pack(ops.Reply(), message.FrameWithUUID(msg.Metadata.UUID), message.FrameIsRPC())
	if err != nil {
		return err
	}
	if err := s.writeMessage(cs, responseMsg); err != nil {
		return err
	}
	return nil
}

func (s *Server) writeMessage(cs connSlot, responseMsg []byte) error {
	cs.l.Lock()
	defer cs.l.Unlock()
	return cs.c.WriteMessage(websocket.BinaryMessage, responseMsg)
}

func (s *Server) RegisterMiddleware(m message.Middleware) {
	s.handlerChainer.AppendMiddleware(m)
}

func (s *Server) RegisterHandler(msg proto.Message, handler message.Handler) {
	fqdn := message.FQDN(msg)
	s.messageRegistry.Register(fqdn, msg)
	s.handlerChainer.AppendHandler(fqdn, handler)
}

func (s *Server) Send(ctx context.Context, msg *goomerang.Message) error {
	ch := make(chan error, 1)
	go func() {
		bytes, err := message.Pack(msg)
		if err != nil {
			ch <- err
			close(ch)
			return
		}
		var errList error
		var count int
		for conn := range s.connRegistry {
			if err := s.writeMessage(s.connRegistry[conn], bytes); err != nil && count < 100 {
				if errors.Is(err, websocket.ErrCloseSent) {
					err = ErrClientDisconnected
				}
				errList = multierror.Append(err, errList)
				count++
			}
		}
		if errList != nil {
			ch <- errList
		}
		close(ch)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

func (s *Server) Run() error {
	s.handlerChainer.PrepareChains()
	if s.cfg.TLSConfig != nil {
		// The "certFile" and "keyFile" params are with "" values, since the server has the certificates already configured.
		if err := s.intServer.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			s.onErrorHook(err)
			return err
		}
		return nil
	}
	if err := s.intServer.ListenAndServe(); err != http.ErrServerClosed {
		s.onErrorHook(err)
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
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
	if s.onCloseHook != nil {
		s.onCloseHook()
	}
	return multiErr
}

func (s *Server) closeConnection(cs connSlot) {
	if err := s.sendClosingSignal(cs); err != nil {
		s.onErrorHook(err)
	}
}

func (s *Server) sendClosingSignal(cs connSlot) error {
	cs.l.Lock()
	defer cs.l.Unlock()
	return cs.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
