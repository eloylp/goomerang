package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func mainHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.wg.Add(1)
		defer s.wg.Done()
		c, err := s.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			s.onErrorHook(err)
			return
		}
		cs := addConnection(s, c)
		defer removeConnection(s, cs)
		defer func() {
			if err := cs.sendCloseSignal(); err != nil {
				s.onErrorHook(err)
			}
			if err := cs.close(); err != nil {
				s.onErrorHook(err)
			}
		}()

		messageReader := readMessages(s, cs)

		for {
			select {
			case <-s.ctx.Done():
				return
			case msg, ok := <-messageReader:
				if !ok { // Channel closed in the receiver, we abort this handler and start defer calling.
					return
				}
				if msg.mType != websocket.BinaryMessage {
					s.onErrorHook(fmt.Errorf("server: cannot process websocket frame type %v", msg.mType))
					continue
				}
				s.workerPool.Add() // Will block till more processing slots are available.
				go func() {
					if err := s.processMessage(cs, msg.data, &stdSender{cs}); err != nil {
						s.onErrorHook(err)
					}
					s.workerPool.Done()
				}()
			}
		}
	}
}
