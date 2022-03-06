package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

type connSlot struct {
	l *sync.Mutex
	c *websocket.Conn
}

func (cs *connSlot) write(msg []byte) error {
	cs.l.Lock()
	defer cs.l.Unlock()
	return cs.c.WriteMessage(websocket.BinaryMessage, msg)
}

func (cs *connSlot) sendCloseSignal() error {
	cs.l.Lock()
	defer cs.l.Unlock()
	return cs.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (cs *connSlot) close() error {
	cs.l.Lock()
	defer cs.l.Unlock()
	return cs.c.Close()
}

type receivedMessage struct {
	mType int
	data  []byte
}
