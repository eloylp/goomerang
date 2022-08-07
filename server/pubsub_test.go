package server

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/conn"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/message"
)

func TestPubSubEngine(t *testing.T) {
	msg := message.New().SetPayload(&protos.MessageV1{})

	t.Run("publication", func(t *testing.T) {
		cg := newPubSubEngine()
		s1 := successfulSender()
		s2 := successfulSender()

		cg.subscribe("topic.a", s1)
		cg.subscribe("topic.a", s2)

		require.NoError(t, cg.publish("topic.a", msg))

		s1.AssertCalled(t, "Send", msg)
		s2.AssertCalled(t, "Send", msg)
	})

	t.Run("same sender, different topics", func(t *testing.T) {
		cg := newPubSubEngine()
		s1 := successfulSender()

		cg.subscribe("topic.a", s1)
		cg.subscribe("topic.b", s1)

		require.NoError(t, cg.publish("topic.a", msg))
		require.NoError(t, cg.publish("topic.b", msg))

		s1.AssertNumberOfCalls(t, "Send", 2)
	})

	t.Run("publication on different topics", func(t *testing.T) {
		cg := newPubSubEngine()
		s1 := successfulSender()
		s2 := successfulSender()

		cg.subscribe("topic.a", s1)
		cg.subscribe("topic.b", s2)

		require.NoError(t, cg.publish("topic.a", msg))
		require.NoError(t, cg.publish("topic.b", msg))

		s1.AssertCalled(t, "Send", msg)
		s2.AssertCalled(t, "Send", msg)
	})

	t.Run("isolates topics", func(t *testing.T) {
		cg := newPubSubEngine()
		s1 := successfulSender()

		cg.subscribe("topic.a", s1)

		require.NoError(t, cg.publish("topic.b", msg))
		s1.AssertNotCalled(t, "Send", msg)
	})

	t.Run("unsubscribe", func(t *testing.T) {
		cg := newPubSubEngine()
		s1 := successfulSender()
		s2 := successfulSender()

		cg.subscribe("topic.a", s1)
		cg.subscribe("topic.a", s2)
		cg.unsubscribe("topic.a", s2.ConnSlot())

		require.NoError(t, cg.publish("topic.a", msg))
		s1.AssertCalled(t, "Send", msg)
		s2.AssertNotCalled(t, "Send", mock.Anything)
	})

	t.Run("unsubscribe when not subscribed", func(t *testing.T) {
		cg := newPubSubEngine()

		s := &senderMock{}
		s.On("ConnSlot").Return(&conn.Slot{})

		cg.unsubscribe("topic.a", s.ConnSlot())

		s.AssertNotCalled(t, "Send", mock.Anything)
		s.AssertCalled(t, "ConnSlot")
	})

	t.Run("unsubscribeAll unsubscribes from all topics", func(t *testing.T) {
		cg := newPubSubEngine()
		s1 := successfulSender()

		cg.subscribe("topic.a", s1)
		cg.subscribe("topic.b", s1)

		cg.unsubscribeAll(s1.ConnSlot())

		require.NoError(t, cg.publish("topic.a", msg))
		require.NoError(t, cg.publish("topic.b", msg))

		s1.AssertNumberOfCalls(t, "Send", 0)
	})

	t.Run("publication on empty topic", func(t *testing.T) {
		cg := newPubSubEngine()
		msg := message.New().SetPayload(&protos.MessageV1{})
		require.NoError(t, cg.publish("topic.a", msg))
	})

	t.Run("max errors in 100", func(t *testing.T) {
		cg := newPubSubEngine()

		for i := 0; i < 200; i++ {
			s := &senderMock{}
			s.On("ConnSlot").Return(&conn.Slot{})
			s.On("Send", mock.Anything).Return(0, errors.New("connection error"))

			cg.subscribe("topic.a", s)
		}

		err := cg.publish("topic.a", msg)
		require.Error(t, err)
		multiErr := err.(*multierror.Error)
		assert.Equal(t, 100, multiErr.Len())
	})
}

func successfulSender() *senderMock {
	s1 := &senderMock{}
	s1.On("ConnSlot").Return(&conn.Slot{})
	s1.On("Send", mock.Anything).Return(10, nil)
	return s1
}

var err error

func BenchmarkPubSubEngine(b *testing.B) {
	cg := newPubSubEngine()
	for i := 0; i < 70000; i++ {
		cg.subscribe("topic.a", newDumbSender())
	}

	msg := message.New().SetPayload(&protos.MessageV1{})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = cg.publish("topic.a", msg)
	}
}

type senderMock struct {
	mock.Mock
}

func (s *senderMock) Send(msg *message.Message) (payloadSize int, err error) {
	args := s.Called(msg)
	return args.Int(0), args.Error(1)
}

func (s *senderMock) ConnSlot() *conn.Slot {
	args := s.Called()
	return args.Get(0).(*conn.Slot)
}

type dumbSender struct {
	cs *conn.Slot
}

func newDumbSender() *dumbSender {
	return &dumbSender{
		cs: &conn.Slot{},
	}
}

func (s *dumbSender) Send(msg *message.Message) (payloadSize int, err error) {
	return 10, nil
}

func (s *dumbSender) ConnSlot() *conn.Slot {
	return s.cs
}
