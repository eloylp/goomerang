package server

import (
	"errors"
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

func addConnection(s *Server, c *websocket.Conn) connSlot {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	slot := connSlot{
		l: &sync.Mutex{},
		c: c,
	}
	s.connRegistry[c] = slot
	return slot
}

func removeConnection(s *Server, c connSlot) {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	delete(s.connRegistry, c.c)
}

func readMessages(s *Server, cs connSlot) chan *receivedMessage {
	ch := make(chan *receivedMessage)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(ch)
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				messageType, data, err := cs.c.ReadMessage()
				if err != nil {
					var closeErr *websocket.CloseError
					if errors.As(err, &closeErr) {
						if closeErr.Code == websocket.CloseNormalClosure {
							return // will trigger normal connection close at handler
						}
					}
					s.onErrorHook(err)
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
