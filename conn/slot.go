package conn

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Slot struct {
	l             *sync.Mutex
	c             *websocket.Conn
	receivedClose chan struct{}
}

func NewSlot(c *websocket.Conn) *Slot {
	return &Slot{
		c:             c,
		l:             &sync.Mutex{},
		receivedClose: make(chan struct{}, 1),
	}
}

func (s *Slot) Conn() *websocket.Conn {
	return s.c
}

func (s *Slot) Write(msg []byte) error {
	s.l.Lock()
	defer s.l.Unlock()
	return s.c.WriteMessage(websocket.BinaryMessage, msg)
}

func (s *Slot) WriteRaw(messageType int, msg []byte) error {
	s.l.Lock()
	defer s.l.Unlock()
	return s.c.WriteMessage(messageType, msg)
}

func (s *Slot) SendCloseSignal() error {
	s.l.Lock()
	defer s.l.Unlock()
	err := s.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// This error is ignored because we suspect once the underlying library receives the close frame,
	// it tries to prevent sending anything more. But we need to send back the close frame when we
	// receive it from the server, in order to accomplish the closing handshake.
	if err == websocket.ErrCloseSent {
		return nil
	}
	return err
}

func (s *Slot) Close() error {
	s.l.Lock()
	defer s.l.Unlock()
	return s.c.Close()
}

func (s *Slot) SetReceivedClose() {
	s.receivedClose <- struct{}{}
}

func (s *Slot) WaitReceivedClose() error {
	ctx, cancl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancl()
	select {
	case <-ctx.Done():
		return fmt.Errorf("connection spend more than 5 seconds to send close. Continuing anyway: %v", ctx.Err())
	case <-s.receivedClose:
		return nil
	}
}
