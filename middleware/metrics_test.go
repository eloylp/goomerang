package middleware_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/message"
	clientMetrics "go.eloylp.dev/goomerang/metrics/client"
	serverMetrics "go.eloylp.dev/goomerang/metrics/server"
	"go.eloylp.dev/goomerang/middleware"
)

func TestServerMetrics(t *testing.T) {
	m, err := middleware.PromHistograms(middleware.PromConfig{
		ReceivedMessageSize:   serverMetrics.ReceivedMessageSize,
		MessageProcessingTime: serverMetrics.MessageProcessingTime,
		SentMessageSize:       serverMetrics.SentMessageSize,
	})
	require.NoError(t, err)

	h := message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_, _ = s.Send(context.Background(), msg)
	})
	sender := &fakeSender{}
	msg := &message.Message{
		Metadata: &message.Metadata{
			PayloadSize: 10,
		},
		Payload: &testMessages.MessageV1{},
	}

	m(h).Handle(sender, msg)

	AssertMetricsHandler(t, promhttp.Handler(), "server")
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
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_size_bytes_bucket{le="10"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_size_bytes_sum 10`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_size_bytes_count 1`, system))

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_sent_message_size_bytes_bucket{le="+Inf"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_sent_message_size_bytes_sum 20`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_sent_message_size_bytes_count 1`, system))
}

type fakeSender struct{}

func (f *fakeSender) Send(_ context.Context, _ *message.Message) (payloadSize int, err error) {
	return 20, nil
}

func TestClientMetrics(t *testing.T) {
	m, err := middleware.PromHistograms(middleware.PromConfig{
		ReceivedMessageSize:   clientMetrics.ReceivedMessageSize,
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
		SentMessageSize:       clientMetrics.SentMessageSize,
	})
	require.NoError(t, err)

	h := message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_, _ = s.Send(context.Background(), msg)
	})
	sender := &fakeSender{}
	msg := &message.Message{
		Metadata: &message.Metadata{
			PayloadSize: 10,
		},
		Payload: &testMessages.MessageV1{},
	}
	m(h).Handle(sender, msg)
	AssertMetricsHandler(t, promhttp.Handler(), "client")
}

func TestPromMiddlewareDoesValidations(t *testing.T) {
	_, err := middleware.PromHistograms(middleware.PromConfig{})
	require.EqualError(t, err, "metrics middleware config: validation error: validate: receivedMessageSize be non nil")
}

func TestPromConfig_Validate_NoReceivedMessageSize(t *testing.T) {
	c := middleware.PromConfig{
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
		SentMessageSize:       clientMetrics.SentMessageSize,
	}
	assert.EqualError(t, c.Validate(), "validate: receivedMessageSize be non nil")
}

func TestPromConfig_Validate_MessageProcessingTime(t *testing.T) {
	c := middleware.PromConfig{
		ReceivedMessageSize: clientMetrics.ReceivedMessageSize,
		SentMessageSize:     clientMetrics.SentMessageSize,
	}
	assert.EqualError(t, c.Validate(), "validate: MessageProcessingTime be non nil")
}

func TestPromConfig_Validate_SentMessageSize(t *testing.T) {
	c := middleware.PromConfig{
		ReceivedMessageSize:   clientMetrics.ReceivedMessageSize,
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
	}
	assert.EqualError(t, c.Validate(), "validate: SentMessageSize be non nil")
}
