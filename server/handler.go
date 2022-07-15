package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	"go.eloylp.dev/goomerang/ws"
)

func mainHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.wg.Add(1)
		defer s.wg.Done()
		if s.status() != ws.StatusRunning {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("server: not running"))
			return
		}
		c, err := s.wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			s.hooks.ExecOnError(err)
			return
		}
		cs := addConnection(s, c)
		defer removeConnection(s, cs)
		defer func() {
			if err := cs.close(); err != nil {
				s.hooks.ExecOnError(err)
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
								s.hooks.ExecOnError(err)
							}
						}
						return // will trigger normal connection close at handler, as channel (ch) will be closed.
					}
					if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
						s.hooks.ExecOnError(err)
						return
					}
					s.hooks.ExecOnError(err)
					continue
				}
				if messageType != websocket.BinaryMessage {
					s.hooks.ExecOnError(fmt.Errorf("server: cannot process websocket frame kind %v", messageType))
					continue
				}
				if s.cfg.MaxConcurrency <= 1 {
					if err := s.processMessage(cs, data, &stdSender{cs, s}); err != nil {
						s.hooks.ExecOnError(err)
					}
					continue
				}
				s.workerPool.Add() // Will block till more processing slots are available.
				s.hooks.ExecOnWorkerStart()
				go func() {
					defer s.hooks.ExecOnWorkerEnd()
					defer s.workerPool.Done()
					if err := s.processMessage(cs, data, &stdSender{cs, s}); err != nil {
						s.hooks.ExecOnError(err)
					}
				}()
			}
		}
	}
}
