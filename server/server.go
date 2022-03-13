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

	"go.eloylp.dev/goomerang/internal/conc"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
)

type Server struct {
	intServer       *http.Server
	connRegistry    map[*websocket.Conn]connSlot
	serverL         *sync.RWMutex
	wg              *sync.WaitGroup
	ctx             context.Context
	cancl           context.CancelFunc
	wsUpgrader      *websocket.Upgrader
	handlerChainer  *messaging.HandlerChainer
	messageRegistry messaging.Registry
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
		handlerChainer:  messaging.NewHandlerChainer(),
		messageRegistry: messaging.Registry{},
		connRegistry:    map[*websocket.Conn]connSlot{},
		serverL:         &sync.RWMutex{},
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

func (s *Server) RegisterMiddleware(m message.Middleware) {
	s.handlerChainer.AppendMiddleware(m)
}

func (s *Server) RegisterHandler(msg proto.Message, handler message.Handler) {
	fqdn := messaging.FQDN(msg)
	s.messageRegistry.Register(fqdn, msg)
	s.handlerChainer.AppendHandler(fqdn, handler)
}

func (s *Server) Send(ctx context.Context, msg *message.Message) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		bytes, err := messaging.Pack(msg)
		if err != nil {
			ch <- err
			return
		}
		var errList error
		var count int
		s.serverL.RLock()
		defer s.serverL.RUnlock()
		for conn := range s.connRegistry {
			cs := s.connRegistry[conn]
			if err := cs.write(bytes); err != nil && count < 100 {
				if errors.Is(err, websocket.ErrCloseSent) {
					err = ErrClientDisconnected
				}
				errList = multierror.Append(err, errList)
				count++
			}
		}
		ch <- errList
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
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		s.cancl()
		var multiErr error
		if err := s.intServer.Shutdown(ctx); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		s.workerPool.Wait() // Wait for in flight user handlers
		s.wg.Wait()         // Wait for in flight server handlers
		if s.onCloseHook != nil {
			s.onCloseHook()
		}
		ch <- multiErr
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

func (s *Server) processMessage(cs connSlot, data []byte, sOpts message.Sender) (err error) {
	frame, err := messaging.UnPack(data)
	if err != nil {
		return err
	}

	msg, err := messaging.FromFrame(frame, s.messageRegistry)
	if err != nil {
		return err
	}
	handler, err := s.handlerChainer.Handler(msg.Metadata.Type)
	if err != nil {
		return err
	}
	if msg.Metadata.IsRPC {
		if err := doRPC(handler, cs, msg); err != nil {
			return err
		}
		return nil
	}
	handler.Handle(sOpts, msg)
	return nil
}
