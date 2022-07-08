package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/conc"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/ws"
)

// Client holds are the client subsystems and dependencies.
type Client struct {
	cfg               *Cfg
	ServerURL         url.URL
	handlerChainer    *messaging.HandlerChainer
	messageRegistry   message.Registry
	writeLock         *sync.Mutex
	wg                *sync.WaitGroup
	conn              *websocket.Conn
	dialer            *websocket.Dialer
	ctx               context.Context
	cancl             context.CancelFunc
	hooks             *hooks
	requestRegistry   *requestRegistry
	workerPool        *conc.WorkerPool
	currentStatus     uint32
	chCloseWait       chan struct{}
	heartbeatInterval time.Duration
}

// New creates a new client, review available options.
func New(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	wp, err := conc.NewWorkerPool(cfg.MaxConcurrency)
	if err != nil {
		return nil, fmt.Errorf("goomerang client: %w", err)
	}
	c := &Client{
		ServerURL:         serverURL(cfg),
		cfg:               cfg,
		hooks:             cfg.hooks,
		heartbeatInterval: cfg.HeartbeatInterval,
		dialer: &websocket.Dialer{
			Proxy:             http.ProxyFromEnvironment,
			TLSClientConfig:   cfg.TLSConfig,
			HandshakeTimeout:  cfg.HandshakeTimeout,
			ReadBufferSize:    cfg.ReadBufferSize,
			WriteBufferSize:   cfg.WriteBufferSize,
			EnableCompression: cfg.EnableCompression,
		},
		writeLock:       &sync.Mutex{},
		wg:              &sync.WaitGroup{},
		handlerChainer:  messaging.NewHandlerChainer(),
		messageRegistry: message.Registry{},
		requestRegistry: newRegistry(),
		workerPool:      wp,
		chCloseWait:     make(chan struct{}, 1),
	}
	c.hooks.ExecOnConfiguration(cfg)
	c.setStatus(ws.StatusNew)
	return c, nil
}

// Connect tries a connection to the specified server.
// This method can be called after a client shutdown again
// in order to retry the connection.
func (c *Client) Connect(ctx context.Context) error {
	if c.status() != ws.StatusNew && c.status() != ws.StatusClosed {
		return ErrAlreadyRunning
	}
	c.ctx, c.cancl = context.WithCancel(context.Background())
	c.handlerChainer.PrepareChains()
	conn, resp, err := c.dialer.DialContext(ctx, c.ServerURL.String(), nil)
	if err != nil {
		return fmt.Errorf("connect: %v", err)
	}
	defer resp.Body.Close()
	c.conn = conn
	c.wg.Add(1)
	go c.receiver()
	c.wg.Add(1)
	go c.heartbeat()
	c.setStatus(ws.StatusRunning)
	return nil
}

func (c *Client) receiver() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			messageType, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					c.receivedCloseFromServer()
					if c.status() == ws.StatusClosing {
						return
					}
					go func() {
						if err := c.close(context.Background(), false); err != nil {
							c.hooks.ExecOnError(err)
						}
					}()
					return
				}
				if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					c.setStatus(ws.StatusClosing)
					go c.forceClose()
				}
				c.hooks.ExecOnError(err)
				return
			}
			if messageType != websocket.BinaryMessage {
				c.hooks.ExecOnError(fmt.Errorf("protocol: unexpected message kind %v", messageType))
				continue
			}
			if c.cfg.MaxConcurrency <= 1 {
				if err := c.processMessage(data); err != nil {
					c.hooks.ExecOnError(err)
				}
				continue
			}
			c.workerPool.Add()
			c.hooks.ExecOnWorkerStart()
			go func() {
				defer c.hooks.ExecOnWorkerEnd()
				defer c.workerPool.Done()
				if err := c.processMessage(data); err != nil {
					c.hooks.ExecOnError(err)
				}
			}()
		}
	}
}

func (c *Client) forceClose() {
	c.cancl()
	c.wg.Wait()
	c.setStatus(ws.StatusClosed)
	c.hooks.ExecOnclose()
}

func (c *Client) processMessage(data []byte) (err error) {
	frame, err := messaging.UnPack(data)
	if err != nil {
		return err
	}
	msg, err := messaging.FromFrame(frame, c.messageRegistry)
	if err != nil {
		return err
	}
	if msg.Metadata.IsSync {
		if err := c.receiveSync(msg); err != nil {
			return err
		}
		return nil
	}
	handler, err := c.handlerChainer.Handler(msg.Metadata.Kind)
	if err != nil {
		return err
	}
	handler.Handle(c, msg)
	return nil
}

// Send will write the provided *message.Message to the output buffer.
// This method is completely asynchronous. In case the client has closed
// the connection due to other reasons, ErrNotRunning will be returned.
//
// If successful, it will return the payload size in bytes. This payload size
// does not include the goomerang base message (which is around 12 bytes) nor
// the headers.
func (c *Client) Send(msg *message.Message) (payloadSize int, err error) {
	if c.status() != ws.StatusRunning {
		return 0, ErrNotRunning
	}
	var data []byte
	payloadSize, data, err = messaging.Pack(msg)
	if err != nil {
		return
	}
	if err = c.writeMessage(data); err != nil {
		return payloadSize, fmt.Errorf("send: %v", err)
	}
	return
}

// SendSync will send a message to the server and wait for a reply. If
// the provided context is canceled, this function will return immediately
// and the reply message will be lost.
//
// If successful, it will return the payload size in bytes. This payload size
// does not include the goomerang base message (which is around 12 bytes) nor
// the headers.
//
// ErrNotRunning error will be returned in case the client was closed for any reason.
func (c *Client) SendSync(ctx context.Context, msg *message.Message) (payloadSize int, response *message.Message, err error) {
	if c.status() != ws.StatusRunning {
		return 0, nil, ErrNotRunning
	}
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		UUID := uuid.New().String()
		var data []byte
		payloadSize, data, err = messaging.Pack(msg, messaging.FrameWithUUID(UUID), messaging.FrameIsSync())
		if err != nil {
			return
		}
		c.requestRegistry.createListener(UUID)
		if err = c.writeMessage(data); err != nil {
			err = fmt.Errorf("sendSync: %v", err)
			return
		}
		response, err = c.requestRegistry.resultFor(ctx, UUID)
	}()
	select {
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	case <-ch:
		return
	}
}

// Close will initiate the graceful shutdown procedure for this
// client. It will work in a best effort way. In case of errors,
// they are going to be collected and returned as a multi-error
// type, with the hope the maximum number of closing actions are performed.
func (c *Client) Close(ctx context.Context) (err error) {
	return c.close(ctx, true)
}

func (c *Client) close(ctx context.Context, isInitiator bool) (err error) {
	if c.status() != ws.StatusRunning {
		return ErrNotRunning
	}
	c.setStatus(ws.StatusClosing)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		defer c.hooks.ExecOnclose()
		defer c.setStatus(ws.StatusClosed)
		errList := multierror.Append(nil, nil)
		if err := c.sendClosingSignal(); err != nil {
			errList = multierror.Append(nil, fmt.Errorf("close: %v", err))
		}
		var serverLooksUnresponsive bool
		if isInitiator {
			// Wait for server shutdown handshake, not forever of course.
			if err := c.waitForServerCloseReply(); err != nil {
				serverLooksUnresponsive = true
				errList = multierror.Append(nil, fmt.Errorf("close: %v", err))
			}
		}
		c.cancl()
		c.workerPool.Wait() // Wait for in flight user handlers
		c.wg.Wait()         // Wait for in flight client handlers

		// Client should never close connections : https://datatracker.ietf.org/doc/html/rfc6455#section-7.1.1
		// Of course, if server did not reply in the close handshake, we ensure the connection close.
		if serverLooksUnresponsive {
			if err := c.conn.Close(); err != nil {
				errList = multierror.Append(errList, fmt.Errorf("close: %v", err))
			}
		}
		err = errList.ErrorOrNil()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return
	}
}

// Middleware registers a middleware in the client. It
// will panic if the client its already running.
func (c *Client) Middleware(m message.Middleware) {
	c.handlerChainer.AppendMiddleware(m)
}

// Handle registers a message handler in the client. It
// will panic if the client its already running.
func (c *Client) Handle(msg proto.Message, h message.Handler) {
	fqdn := messaging.FQDN(msg)
	c.messageRegistry.Register(fqdn, msg)
	c.handlerChainer.AppendHandler(fqdn, h)
}

// RegisterMessage will make the client aware of a specific kind of
// protocol buffer message. This is specially needed when
// the user sends messages with methods like SendSync(),
// as the client needs to know how to decode the incoming reply.
//
// If the kind of message it's already registered with the Handle()
// method, then the user can omit this registration.
//
// Any kind of protocol buffers message that arrives to the client,
// and It's not registered, will be discarded.
func (c *Client) RegisterMessage(msg proto.Message) {
	c.messageRegistry.Register(messaging.FQDN(msg), msg)
}

func (c *Client) writeMessage(data []byte) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) sendClosingSignal() error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// This error is ignored because we suspect once the underlying library receives the close frame,
	// it tries to prevent sending anything more. But we need to send back the close frame when we
	// receive it from the server, in order to accomplish the closing handshake.
	if err == websocket.ErrCloseSent {
		return nil
	}
	return err
}

func (c *Client) receiveSync(msg *message.Message) error {
	if err := c.requestRegistry.submitResult(msg.Metadata.UUID, msg); err != nil {
		return err
	}
	return nil
}

func (c *Client) setStatus(status uint32) {
	atomic.StoreUint32(&c.currentStatus, status)
	c.hooks.ExecOnStatusChange(status)
}

func (c *Client) status() uint32 {
	return atomic.LoadUint32(&c.currentStatus)
}

func (c *Client) waitForServerCloseReply() error {
	ctx, cancl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancl()
	select {
	case <-ctx.Done():
		return fmt.Errorf("server spent more than 5 seconds to send close. Continuing anyway: %v", ctx.Err())
	case <-c.chCloseWait:
		return nil
	}
}

func (c *Client) receivedCloseFromServer() {
	c.chCloseWait <- struct{}{}
}

func (c *Client) heartbeat() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()
	pingFn := func() (err error) {
		c.writeLock.Lock()
		defer c.writeLock.Unlock()
		if c.status() == ws.StatusRunning {
			err = c.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
		}
		return
	}
	for {
		select {
		case <-ticker.C:
			if err := pingFn(); err != nil {
				c.hooks.ExecOnError(err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}
