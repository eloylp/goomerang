package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/conn"
	"go.eloylp.dev/goomerang/internal/conc"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/internal/messaging/protocol"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/ws"
)

// Server will hold all subsystem and orchestration
// elements.
type Server struct {
	intServer       *http.Server
	connRegistry    map[*websocket.Conn]*conn.Slot
	serverL         *sync.RWMutex
	wg              *sync.WaitGroup
	ctx             context.Context
	cancl           context.CancelFunc
	pubSubEngine    *pubSubEngine
	wsUpgrader      *websocket.Upgrader
	handlerChainer  *messaging.HandlerChainer
	messageRegistry message.Registry
	hooks           *hooks
	cfg             *Cfg
	workerPool      *conc.WorkerPool
	currentStatus   uint32
	chCloseWait     chan struct{}
	listener        net.Listener
}

// New creates a server instance, check all available
// options.
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
		connRegistry: map[*websocket.Conn]*conn.Slot{},
		serverL:      &sync.RWMutex{},
		wg:           &sync.WaitGroup{},
		ctx:          ctx,
		cancl:        cancl,
		pubSubEngine: newPubSubEngine(),
		wsUpgrader: &websocket.Upgrader{
			HandshakeTimeout:  cfg.HandshakeTimeout,
			ReadBufferSize:    cfg.ReadBufferSize,
			WriteBufferSize:   cfg.WriteBufferSize,
			EnableCompression: cfg.EnableCompression,
		},
		handlerChainer:  messaging.NewHandlerChainer(),
		messageRegistry: message.Registry{},
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

// Middleware registers a middleware in the server. It
// will panic if the server its already running.
func (s *Server) Middleware(m message.Middleware) {
	s.handlerChainer.AppendMiddleware(m)
}

// Handle registers a message handler in the server. It
// will panic if the server its already running.
func (s *Server) Handle(msg proto.Message, handler message.Handler) {
	fqdn := messaging.FQDN(msg)
	s.messageRegistry.Register(fqdn, msg)
	s.handlerChainer.AppendHandler(fqdn, handler)
}

// BroadcastResult represents the size and duration
// of each sent message during a broadcast operation.
// It provides information to the caller code.
type BroadcastResult struct {
	Size     int
	Duration time.Duration
}

// Broadcast will try to send the provided message to all
// connected clients.
//
// In case the provided context is canceled, the operation
// could be partially completed. It's recommended to always
// pass a context.WithTimeout().
//
// The send operation will happen in parallel for each client.
//
// In case of errors, it will continue the operation,
// keeping the first 100 ones and returning them in a
// multi-error type.
//
// In case of success, the BroadcastResult return type will provide
// feedback, such as time distribution.
//
// Calling this method it's intended to be thread safe.
func (s *Server) Broadcast(ctx context.Context, msg *message.Message) (brResult []BroadcastResult, err error) {
	if s.status() != ws.StatusRunning {
		return []BroadcastResult{}, ErrNotRunning
	}
	ch := make(chan struct{})
	go func() {
		defer close(ch)

		var data []byte
		var payloadSize int
		payloadSize, data, err = messaging.Pack(msg)

		if err != nil {
			err = fmt.Errorf("broadcast: %v", err)
			return
		}

		const maxErrors = 100
		errs := make([]error, 0, maxErrors)
		var errCount int

		s.serverL.RLock()
		defer s.serverL.RUnlock()

		brResult = make([]BroadcastResult, 0, len(s.connRegistry))

		for c := range s.connRegistry {
			cs := s.connRegistry[c]
			start := time.Now()
			if err := cs.Write(data); err != nil && errCount < maxErrors {
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

// Publish enables the server to send a message to all
// subscribed clients in a topic.
func (s *Server) Publish(topic string, msg *message.Message) error {
	if s.status() != ws.StatusRunning {
		return ErrNotRunning
	}
	if err := s.pubSubEngine.publish(topic, msg); err != nil {
		return fmt.Errorf("publish: %v", err)
	}
	return nil
}

// Run will start the server with all the previously
// provided options.
//
// The server instance it's not reusable. This method
// can be only executed once in the lifecycle of a server.
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
	registerBuiltInHandlers(s)
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

// Router provides access to the underlying HTTP router.
// The routes /ws and /wss are reserved for library
// operations.
func (s *Server) Router() *http.ServeMux {
	if s.status() != ws.StatusNew {
		panic(errors.New("custom handlers can only be implemented before running"))
	}
	return s.intServer.Handler.(*http.ServeMux)
}

// RegisterMessage will make the server aware of a specific kind of
// protocol buffer message. This is specially needed when
// the user sends messages with methods like SendSync(),
// as the client needs to know how to decode the incoming reply.
//
// If the kind of message it's already registered with the Handle()
// method, then the user can omit this registration.
//
// This is specially needed in order to make the server aware
// of the messages published in pub/sub patterns.
func (s *Server) RegisterMessage(msg proto.Message) {
	s.messageRegistry.Register(messaging.FQDN(msg), msg)
}

// Addr provides the listener address. It will
// return an empty string if it's not possible to
// determine the IP information.
func (s *Server) Addr() string {
	if s.status() != ws.StatusRunning {
		return ""
	}
	return s.listener.Addr().String()
}

// Shutdown will start the graceful shutdown of the server.
//
// The call to this function will block till the operation completes or
// the provided context expires.
//
// A closing signal will be sent to all connected clients in parallel. Clients
// should then individually reply to this signal with another closing one. The server
// will wait a grace period of 5 seconds per each connection, before closing them.
//
// This function will also wait for all the inflight message handlers and subsystems.
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

func registerBuiltInHandlers(s *Server) {
	s.Handle(&protocol.BroadcastCmd{}, broadcastCmdHandler(s))
	s.Handle(&protocol.SubscribeCmd{}, subscribeCmdHandler(s.pubSubEngine, s.hooks.ExecOnSubscribe))
	s.Handle(&protocol.PublishCmd{}, publishCmdHandler(s.messageRegistry, s.pubSubEngine, s.hooks.ExecOnPublish, s.hooks.ExecOnError))
	s.Handle(&protocol.UnsubscribeCmd{}, unsubscribeCmdHandler(s.pubSubEngine, s.hooks.ExecOnUnsubscribe))
}

func (s *Server) broadcastClose() {
	s.serverL.RLock()
	defer s.serverL.RUnlock()
	wg := sync.WaitGroup{}
	for _, cs := range s.connRegistry {
		wg.Add(1)
		go func(cs *conn.Slot) {
			defer wg.Done()
			if err := cs.SendCloseSignal(); err != nil {
				s.hooks.ExecOnError(err)
			}
			if err := cs.WaitReceivedClose(); err != nil {
				s.hooks.ExecOnError(err)
			}
		}(cs)
	}
	wg.Wait()
}

func (s *Server) processMessage(cs *conn.Slot, data []byte) (err error) {
	frame, err := messaging.UnPack(data)
	if err != nil {
		return err
	}

	msg, err := messaging.FromFrame(frame, s.messageRegistry)
	if err != nil {
		return err
	}
	handler, err := s.handlerChainer.Handler(msg.Metadata.Kind)
	if err != nil {
		return err
	}
	if msg.Metadata.IsSync {
		handler.Handle(&SyncSender{cs, s.status, msg.Metadata.UUID}, msg)
		return nil
	}
	handler.Handle(&stdSender{cs, s.status}, msg)
	return nil
}

func (s *Server) setStatus(status uint32) {
	atomic.StoreUint32(&s.currentStatus, status)
	s.hooks.ExecOnStatusChange(status)
}

func (s *Server) status() uint32 {
	return atomic.LoadUint32(&s.currentStatus)
}

func (s *Server) Registry() message.Registry {
	return s.messageRegistry
}
