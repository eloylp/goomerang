package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/conc"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/ws"
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
	hooks           *hooks
	cfg             *Cfg
	workerPool      *conc.WorkerPool
	currentStatus   uint32
	chCloseWait     chan struct{}
	listener        net.Listener
}

func New(opts ...Option) (*Server, error) {
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
		intServer: &http.Server{
			TLSConfig:         cfg.TLSConfig,
			ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
			ReadTimeout:       cfg.HTTPReadTimeout,
			WriteTimeout:      cfg.HTTPWriteTimeout,
		},
		connRegistry: map[*websocket.Conn]connSlot{},
		serverL:      &sync.RWMutex{},
		wg:           &sync.WaitGroup{},
		ctx:          ctx,
		cancl:        cancl,
		wsUpgrader: &websocket.Upgrader{
			HandshakeTimeout:  cfg.HandshakeTimeout,
			ReadBufferSize:    cfg.ReadBufferSize,
			WriteBufferSize:   cfg.WriteBufferSize,
			EnableCompression: cfg.EnableCompression,
		},
		handlerChainer:  messaging.NewHandlerChainer(),
		messageRegistry: messaging.Registry{},
		hooks:           cfg.hooks,
		cfg:             cfg,
		workerPool:      wp,
		chCloseWait:     make(chan struct{}, 1),
	}
	mux := http.NewServeMux()
	mux.Handle(endpoint(cfg), mainHandler(s))
	s.intServer.Handler = mux
	s.hooks.ExecOnConfiguration(cfg)
	s.setStatus(ws.StatusNew)
	return s, nil
}

func (s *Server) Middleware(m message.Middleware) {
	s.handlerChainer.AppendMiddleware(m)
}

func (s *Server) Handle(msg proto.Message, handler message.Handler) {
	fqdn := messaging.FQDN(msg)
	s.messageRegistry.Register(fqdn, msg)
	s.handlerChainer.AppendHandler(fqdn, handler)
}

type BroadcastResult struct {
	Size     int
	Duration time.Duration
}

func (s *Server) BroadCast(ctx context.Context, msg *message.Message) (brResult []BroadcastResult, err error) {
	if s.status() != ws.StatusRunning {
		return []BroadcastResult{}, ErrNotRunning
	}
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		var data []byte
		payloadSize, data, err := messaging.Pack(msg)
		if err != nil {
			return
		}
		errs := make([]error, 0, 100)
		var errCount int
		s.serverL.RLock()
		defer s.serverL.RUnlock()
		brResult = make([]BroadcastResult, 0, len(s.connRegistry))
		for conn := range s.connRegistry {
			cs := s.connRegistry[conn]
			start := time.Now()
			if err := cs.write(data); err != nil && errCount < 100 {
				errs = append(errs, fmt.Errorf("broadCast: %v", err))
				errCount++
				continue
			}
			brResult = append(brResult, BroadcastResult{
				Size:     payloadSize,
				Duration: time.Since(start),
			})
		}
		err = multierror.Append(err, errs...).ErrorOrNil()
	}()
	select {
	case <-ctx.Done():
		return []BroadcastResult{}, ctx.Err()
	case <-ch:
		return
	}
}

func (s *Server) Run() (err error) {
	if s.status() == ws.StatusClosed {
		return ErrClosed
	}
	if s.status() == ws.StatusClosing {
		return ErrClosing
	}
	if s.status() == ws.StatusRunning {
		return ErrAlreadyRunning
	}
	s.listener, err = net.Listen("tcp", s.cfg.ListenURL)
	if err != nil {
		return fmt.Errorf("run: %v", err)
	}
	s.setStatus(ws.StatusRunning)
	s.handlerChainer.PrepareChains()

	if s.cfg.TLSConfig != nil {
		// The "certFile" and "keyFile" params are with "" values, since the server has the certificates already configured.
		if err := s.intServer.ServeTLS(s.listener, "", ""); err != http.ErrServerClosed {
			err = fmt.Errorf("run: %v", err)
			return err
		}
		return nil
	}
	if err := s.intServer.Serve(s.listener); err != http.ErrServerClosed {
		err = fmt.Errorf("run: %v", err)
		return err
	}
	return nil
}

func (s *Server) Addr() string {
	if s.status() != ws.StatusRunning {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *Server) Shutdown(ctx context.Context) (err error) {
	if s.status() != ws.StatusRunning {
		return ErrNotRunning
	}
	s.setStatus(ws.StatusClosing)
	defer s.setStatus(ws.StatusClosed)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		defer s.hooks.ExecOnclose()

		s.broadcastClose()
		s.cancl() // This will finish all connections, as will cancel all receiver goroutines. See uses.
		err = s.intServer.Shutdown(ctx)
		s.workerPool.Wait() // Wait for in flight user handlers
		s.wg.Wait()         // Wait for in flight server handlers
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return
	}
}

func (s *Server) broadcastClose() {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	wg := sync.WaitGroup{}
	for _, cs := range s.connRegistry {
		wg.Add(1)
		go func(cs connSlot) {
			defer wg.Done()
			if err := cs.sendCloseSignal(); err != nil {
				s.hooks.ExecOnError(err)
			}
			if err := cs.waitReceivedClose(); err != nil {
				s.hooks.ExecOnError(err)
			}
		}(cs)
	}
	wg.Wait()
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
	if msg.Metadata.IsSync {
		handler.Handle(&SyncSender{cs, msg.Metadata.UUID}, msg)
		return nil
	}
	handler.Handle(sOpts, msg)
	return nil
}

func (s *Server) setStatus(status uint32) {
	atomic.StoreUint32(&s.currentStatus, status)
	s.hooks.ExecOnStatusChange(status)
}

func (s *Server) status() uint32 {
	return atomic.LoadUint32(&s.currentStatus)
}
