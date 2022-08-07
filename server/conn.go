package server

import (
	"github.com/gorilla/websocket"

	"go.eloylp.dev/goomerang/conn"
)

func addConnection(s *Server, c *websocket.Conn) *conn.Slot {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	slot := conn.NewSlot(c)
	s.connRegistry[c] = slot
	return slot
}

func removeConnection(s *Server, cs *conn.Slot) {
	s.serverL.Lock()
	defer s.serverL.Unlock()
	delete(s.connRegistry, cs.Conn())
}
