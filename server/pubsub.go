package server

import (
	"sync"

	"github.com/hashicorp/go-multierror"

	"go.eloylp.dev/goomerang/conn"
	"go.eloylp.dev/goomerang/message"
)

type pubSubEngine struct {
	csMap map[string]map[*conn.Slot]message.Sender
	L     *sync.RWMutex
}

func newPubSubEngine() *pubSubEngine {
	return &pubSubEngine{
		csMap: map[string]map[*conn.Slot]message.Sender{},
		L:     &sync.RWMutex{},
	}
}

func (cm *pubSubEngine) subscribe(topic string, sender message.Sender) {
	cm.L.Lock()
	defer cm.L.Unlock()
	ct, ok := cm.csMap[topic]
	if !ok {
		cm.csMap[topic] = map[*conn.Slot]message.Sender{
			sender.ConnSlot(): sender,
		}
		return
	}
	ct[sender.ConnSlot()] = sender
}

func (cm *pubSubEngine) publish(topic string, msg *message.Message) error {
	cm.L.RLock()
	defer cm.L.RUnlock()
	var multiErr *multierror.Error
	var count int
	for _, sender := range cm.csMap[topic] {
		if _, err := sender.Send(msg); err != nil && count < 100 {
			multiErr = multierror.Append(multiErr, err)
			count++
		}
	}
	return multiErr.ErrorOrNil()
}

func (cm *pubSubEngine) unsubscribe(topic string, cs *conn.Slot) {
	cm.L.Lock()
	defer cm.L.Unlock()
	delete(cm.csMap[topic], cs)
}

func (cm *pubSubEngine) unsubscribeAll(cs *conn.Slot) {
	cm.L.Lock()
	defer cm.L.Unlock()
	for _, m := range cm.csMap {
		delete(m, cs)
	}
}
