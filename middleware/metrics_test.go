package middleware_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/message"
	clientMetrics "go.eloylp.dev/goomerang/metrics/client"
	serverMetrics "go.eloylp.dev/goomerang/metrics/server"
	"go.eloylp.dev/goomerang/middleware"
)

func TestMetrics(t *testing.T) {

	registry := prometheus.NewRegistry()

	server := serverMetrics.NewMetrics(serverMetrics.DefaultConfig())
	server.Register(registry)

	client := clientMetrics.NewMetrics(clientMetrics.DefaultConfig())
	client.Register(registry)

	cases := []struct {
		side   string
		config middleware.PromConfig
	}{
		{
			"server", middleware.PromConfig{
				MessageInflightTime:   server.MessageInflightTime,
				MessageReceivedSize:   server.MessageReceivedSize,
				MessageProcessingTime: server.MessageProcessingTime,
				MessageSentSize:       server.MessageSentSize,
				MessageSentTime:       server.MessageSentTime,
			}},
		{
			"client", middleware.PromConfig{
				MessageInflightTime:   client.MessageInflightTime,
				MessageReceivedSize:   client.MessageReceivedSize,
				MessageProcessingTime: client.MessageProcessingTime,
				MessageSentSize:       client.MessageSentSize,
				MessageSentTime:       client.MessageSentTime,
			}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("testing metrics for %s", c.side), func(t *testing.T) {
			m, err := middleware.PromHistograms(c.config)
			require.NoError(t, err)

			h := message.HandlerFunc(func(s message.Sender, msg *message.Message) {
				msg = message.New().SetPayload(&testMessages.MessageV1{})
				_, _ = s.Send(msg)
			})

			sender := &fakeSender{}
			msg := &message.Message{
				Metadata: message.Metadata{
					PayloadSize: 10,
					Creation:    time.Now().Add(-1 * time.Second),
					Kind:        "goomerang.test.MessageV1",
				},
				Payload: &testMessages.MessageV1{},
			}

			m(h).Handle(sender, msg)

			AssertMetricsHandler(t, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}), c.side)
		})
	}
}

func AssertMetricsHandler(t *testing.T, handler http.Handler, system string) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	data, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_received_inflight_duration_seconds_bucket{kind="goomerang.test.MessageV1",le="2.5"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_received_inflight_duration_seconds_sum{kind="goomerang.test.MessageV1"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_received_inflight_duration_seconds_count{kind="goomerang.test.MessageV1"} 1`, system))

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_received_size_bytes_bucket{kind="goomerang.test.MessageV1",le="10"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_received_size_bytes_sum{kind="goomerang.test.MessageV1"} 10`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_received_size_bytes_count{kind="goomerang.test.MessageV1"} 1`, system))

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_sent_size_bytes_bucket{kind="goomerang.test.MessageV1",le="+Inf"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_sent_size_bytes_sum{kind="goomerang.test.MessageV1"} 20`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_sent_size_bytes_count{kind="goomerang.test.MessageV1"} 1`, system))

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_sent_duration_seconds_bucket{kind="goomerang.test.MessageV1",le="+Inf"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_sent_duration_seconds_sum{kind="goomerang.test.MessageV1"}`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_sent_duration_seconds_count{kind="goomerang.test.MessageV1"} 1`, system))
}

func TestMetricsIfNotSendsMessageBack(t *testing.T) {
	server := serverMetrics.NewMetrics(serverMetrics.DefaultConfig())
	m, err := middleware.PromHistograms(middleware.PromConfig{
		MessageInflightTime:   server.MessageInflightTime,
		MessageReceivedSize:   server.MessageReceivedSize,
		MessageProcessingTime: server.MessageProcessingTime,
		MessageSentSize:       server.MessageSentSize,
		MessageSentTime:       server.MessageSentTime,
	})
	require.NoError(t, err)

	h := message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		msg = message.New().SetPayload(&testMessages.MessageV1{})
	})
	msg := &message.Message{
		Metadata: message.Metadata{
			PayloadSize: 10,
			Creation:    time.Now().Add(-1 * time.Second),
			Kind:        "goomerang.test.MessageV1",
		},
		Payload: &testMessages.MessageV1{},
	}
	assert.NotPanics(t, func() {
		m(h).Handle(&fakeSender{}, msg)
	})
}

type fakeSender struct{}

func (f *fakeSender) Send(msg *message.Message) (payloadSize int, err error) {
	msg.Metadata = message.Metadata{
		Kind: "goomerang.test.MessageV1",
	}
	return 20, nil
}

func TestPromMiddlewareDoesValidations(t *testing.T) {
	_, err := middleware.PromHistograms(middleware.PromConfig{})
	require.EqualError(t, err, "metrics middleware config: validation error: validate: messageInflightTime be non nil")
}

func TestPromConfig_Validate_MessageInflightTime(t *testing.T) {
	m := clientMetrics.NewMetrics(clientMetrics.DefaultConfig())
	c := middleware.PromConfig{
		MessageReceivedSize:   m.MessageReceivedSize,
		MessageProcessingTime: m.MessageProcessingTime,
		MessageSentSize:       m.MessageSentSize,
	}
	assert.EqualError(t, c.Validate(), "validate: messageInflightTime be non nil")
}

func TestPromConfig_Validate_NoReceivedMessageSize(t *testing.T) {
	m := clientMetrics.NewMetrics(clientMetrics.DefaultConfig())
	c := middleware.PromConfig{
		MessageInflightTime:   m.MessageInflightTime,
		MessageProcessingTime: m.MessageProcessingTime,
		MessageSentSize:       m.MessageSentSize,
	}
	assert.EqualError(t, c.Validate(), "validate: receivedMessageSize be non nil")
}

func TestPromConfig_Validate_MessageProcessingTime(t *testing.T) {
	m := clientMetrics.NewMetrics(clientMetrics.DefaultConfig())
	c := middleware.PromConfig{
		MessageInflightTime: m.MessageInflightTime,
		MessageReceivedSize: m.MessageReceivedSize,
		MessageSentSize:     m.MessageSentSize,
	}
	assert.EqualError(t, c.Validate(), "validate: MessageProcessingTime be non nil")
}

func TestPromConfig_Validate_SentMessageSize(t *testing.T) {
	m := clientMetrics.NewMetrics(clientMetrics.DefaultConfig())
	c := middleware.PromConfig{
		MessageInflightTime:   m.MessageInflightTime,
		MessageReceivedSize:   m.MessageReceivedSize,
		MessageProcessingTime: m.MessageProcessingTime,
	}
	assert.EqualError(t, c.Validate(), "validate: MessageSentSize be non nil")
}

func TestPromConfig_Validate_SentMessageTime(t *testing.T) {
	m := clientMetrics.NewMetrics(clientMetrics.DefaultConfig())
	c := middleware.PromConfig{
		MessageInflightTime:   m.MessageInflightTime,
		MessageReceivedSize:   m.MessageReceivedSize,
		MessageProcessingTime: m.MessageProcessingTime,
		MessageSentSize:       m.MessageSentSize,
	}
	assert.EqualError(t, c.Validate(), "validate: MessageSentTime be non nil")
}
