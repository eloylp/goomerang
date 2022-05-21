package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	"go.eloylp.dev/goomerang/internal/ws"
)

func mainHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.wg.Add(1)
		defer s.wg.Done()
		if s.status() != ws.StatusRunning {
			w.WriteHeader(http.StatusUnavailableForLegalReasons)
			_, _ = w.Write([]byte("server: not running"))
			return
		}
		c, err := s.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			s.onErrorHook(err)
			return
		}
		cs := addConnection(s, c)
		defer removeConnection(s, cs)
		defer func() {
			if err := cs.close(); err != nil {
				s.onErrorHook(err)
			}
		}()

		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				messageType, data, err := cs.c.ReadMessage()
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						cs.setReceivedClose()
						// If the close was not initiated by this server, we just ack
						// the close handshake to the client.
						if s.status() != ws.StatusClosing {
							if err := cs.sendCloseSignal(); err != nil {
								s.onErrorHook(err)
							}
						}
						return // will trigger normal connection close at handler, as channel (ch) will be closed.
					}
					if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
						s.onErrorHook(err)
						return
					}
					s.onErrorHook(err)
					continue
				}
				if messageType != websocket.BinaryMessage {
					s.onErrorHook(fmt.Errorf("server: cannot process websocket frame type %v", messageType))
					continue
				}
				s.workerPool.Add() // Will block till more processing slots are available.
				go func() {
					defer s.workerPool.Done()
					if err := s.processMessage(cs, data, &stdSender{cs, s}); err != nil {
						s.onErrorHook(err)
					}
				}()
			}
		}
	}
}
