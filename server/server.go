package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/internal/engine"
	"go.eloylp.dev/goomerang/internal/message"
	"go.eloylp.dev/goomerang/internal/message/protocol"
)

type Handler func(ops Sender, msg proto.Message) *HandlerError

type Server struct {
	intServer              *http.Server
	connTrack              map[*websocket.Conn]struct{}
	writeLock              *sync.Mutex
	wg                     *sync.WaitGroup
	ctx                    context.Context
	cancl                  context.CancelFunc
	wsUpgrader             *websocket.Upgrader
	handlerRegistry        engine.AppendableRegistry
	messageRegistry        message.Registry
	onErrorHook            func(err error)
	onCloseHook            func()
	onMessageProcessedHook timedHook
	onMessageReceivedHook  timedHook
	cfg                    *Config
	workerPool             *goomerang.WorkerPool
}

func NewServer(opts ...Option) (*Server, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	wp, err := goomerang.NewWorkerPool(cfg.MaxConcurrency)
	if err != nil {
		return nil, fmt.Errorf("goomerang server: %w", err)
	}
	ctx, cancl := context.WithCancel(context.Background())
	s := &Server{
		wsUpgrader: &websocket.Upgrader{},
		intServer: &http.Server{
			Addr:      cfg.ListenURL,
			TLSConfig: cfg.TLSConfig,
		},
		onErrorHook:            cfg.OnErrorHook,
		onCloseHook:            cfg.OnCloseHook,
		onMessageProcessedHook: cfg.OnMessageProcessedHook,
		onMessageReceivedHook:  cfg.OnMessageReceivedHook,
		handlerRegistry:        engine.AppendableRegistry{},
		messageRegistry:        message.Registry{},
		connTrack:              map[*websocket.Conn]struct{}{},
		writeLock:              &sync.Mutex{},
		cancl:                  cancl,
		wg:                     &sync.WaitGroup{},
		ctx:                    ctx,
		cfg:                    cfg,
		workerPool:             wp,
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
		s.addConnection(c)
		sOpts := &immediateSender{s, c, s.writeLock}
		defer s.removeConnection(c)
		defer s.closeConnection(c)

		messageReader := s.readMessages(c)

		for {
			select {
			case <-s.ctx.Done():
				s.onErrorHook(s.ctx.Err())
				return
			case msg, ok := <-messageReader:
				if !ok {
					return
				}
				if msg.mType != websocket.BinaryMessage {
					s.onErrorHook(fmt.Errorf("server: cannot process websocket frame type %v", msg.mType))
					return
				}
				s.workerPool.Add() // Will block till more processing slots are available.
				go func() {
					if err := s.processMessage(c, msg.data, sOpts); err != nil {
						s.onErrorHook(err)
					}
					s.workerPool.Done()
				}()
			}
		}
	}
}

func (s *Server) addConnection(c *websocket.Conn) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	s.connTrack[c] = struct{}{}
}

func (s *Server) removeConnection(c *websocket.Conn) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	delete(s.connTrack, c)
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
	}()
	return ch
}

type receivedMessage struct {
	mType int
	data  []byte
}

func (s *Server) processMessage(c *websocket.Conn, data []byte, sOpts Sender) error {
	start := time.Now()
	frame, err := message.UnPack(data)
	if err != nil {
		return err
	}
	if s.onMessageReceivedHook != nil {
		s.onMessageReceivedHook(frame.Type, time.Since(frame.Creation.AsTime()))
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
		if s.onMessageProcessedHook != nil {
			s.onMessageProcessedHook(frame.Type, time.Since(start))
		}
		return nil
	}
	s.processAsync(handlers, sOpts, msg)
	if s.onMessageProcessedHook != nil {
		s.onMessageProcessedHook(frame.Type, time.Since(start))
	}
	return nil
}

func (s *Server) doRPC(handlers []interface{}, c *websocket.Conn, frameUUID string, msg proto.Message) error {
	ops := &bufferedSender{}
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

func (s *Server) nolockWriteMessage(c *websocket.Conn, responseMsg []byte) error {
	return c.WriteMessage(websocket.BinaryMessage, responseMsg)
}

func (s *Server) writeMessage(c *websocket.Conn, responseMsg []byte) error {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	return c.WriteMessage(websocket.BinaryMessage, responseMsg)
}

func (s *Server) processAsync(handlers []interface{}, ops Sender, msg proto.Message) {
	for _, h := range handlers {
		if err := h.(Handler)(ops, msg); err != nil {
			s.onErrorHook(fmt.Errorf("server handler err: %w", err))
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
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
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
		for conn := range s.connTrack {
			if err := s.nolockWriteMessage(conn, bytes); err != nil && count < 100 {
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

func (s *Server) closeConnection(c *websocket.Conn) {
	if err := s.sendClosingSignal(c); err != nil {
		s.onErrorHook(err)
	}
}

func (s *Server) sendClosingSignal(conn *websocket.Conn) error {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
