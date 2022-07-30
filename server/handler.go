package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/internal/messaging/protocol"
	"go.eloylp.dev/goomerang/message"
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
			s.pubSubEngine.unsubscribeAll(cs)
			if err := cs.Close(); err != nil {
				s.hooks.ExecOnError(err)
			}
		}()
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				messageType, data, err := cs.Conn().ReadMessage()
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						cs.SetReceivedClose()
						// If the close was not initiated by this server, we just ack
						// the close handshake to the client.
						if s.status() != ws.StatusClosing {
							if err := cs.SendCloseSignal(); err != nil {
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
					if err := s.processMessage(cs, data); err != nil {
						s.hooks.ExecOnError(err)
					}
					continue
				}
				s.workerPool.Add() // Will block till more processing slots are available.
				s.hooks.ExecOnWorkerStart()
				go func() {
					defer s.hooks.ExecOnWorkerEnd()
					defer s.workerPool.Done()
					if err := s.processMessage(cs, data); err != nil {
						s.hooks.ExecOnError(err)
					}
				}()
			}
		}
	}
}

func subscribeCmdHandler(pse *pubSubEngine, hook func(topic string)) message.Handler {
	return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		subsCmd := msg.Payload.(*protocol.SubscribeCmd)
		pse.subscribe(subsCmd.Topic, s)
		hook(subsCmd.Topic)
	})
}

func unsubscribeCmdHandler(pse *pubSubEngine, hook func(topic string)) message.Handler {
	return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		unsubsCmd := msg.Payload.(*protocol.UnsubscribeCmd)
		pse.unsubscribe(unsubsCmd.Topic, s.ConnSlot())
		hook(unsubsCmd.Topic)
	})
}

func publishCmdHandler(mr message.Registry, pse *pubSubEngine,
	hook func(topic, fqdn string), onErrorHook func(err error)) message.Handler {
	return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		pubCmd := msg.Payload.(*protocol.PublishCmd)
		origMsg, err := messaging.MessageFromPublish(mr, msg)
		if err != nil {
			onErrorHook(err)
			return
		}
		if err := pse.publish(pubCmd.Topic, origMsg); err != nil {
			onErrorHook(err)
		}
		hook(pubCmd.Topic, origMsg.Metadata.Kind)
	})
}
