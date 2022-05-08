package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type connSlot struct {
	l             *sync.Mutex
	c             *websocket.Conn
	receivedClose chan struct{}
}

func (cs *connSlot) write(msg []byte) error {
	cs.l.Lock()
	defer cs.l.Unlock()
	return cs.c.WriteMessage(websocket.BinaryMessage, msg)
}

func (cs *connSlot) sendCloseSignal() error {
	cs.l.Lock()
	defer cs.l.Unlock()
	err := cs.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err == websocket.ErrCloseSent {
		return nil
	}
	return err
}

func (cs *connSlot) close() error {
	cs.l.Lock()
	defer cs.l.Unlock()
	return cs.c.Close()
}

func (cs *connSlot) setReceivedClose() {
	cs.receivedClose <- struct{}{}
}

func (cs *connSlot) waitReceivedClose() error {
	ctx, cancl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancl()
	select {
	case <-ctx.Done():
		return fmt.Errorf("client connection spend more than 5 seconds to send close. Continuing anyway: %v", ctx.Err())
	case <-cs.receivedClose:
		return nil
	}
}

func addConnection(s *Server, c *websocket.Conn) connSlot {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	slot := connSlot{
		l:             &sync.Mutex{},
		c:             c,
		receivedClose: make(chan struct{}, 1),
	}
	s.connRegistry[c] = slot
	return slot
}

func removeConnection(s *Server, c connSlot) {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	delete(s.connRegistry, c.c)
}
