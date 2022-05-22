package middleware_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	cases := []struct {
		side   string
		config middleware.PromConfig
	}{
		{
			"server", middleware.PromConfig{
				MessageInflightTime:   serverMetrics.MessageInflightTime,
				ReceivedMessageSize:   serverMetrics.ReceivedMessageSize,
				MessageProcessingTime: serverMetrics.MessageProcessingTime,
				SentMessageSize:       serverMetrics.SentMessageSize,
				SentMessageTime:       serverMetrics.SentMessageTime,
			}},
		{
			"client", middleware.PromConfig{
				MessageInflightTime:   clientMetrics.MessageInflightTime,
				ReceivedMessageSize:   clientMetrics.ReceivedMessageSize,
				MessageProcessingTime: clientMetrics.MessageProcessingTime,
				SentMessageSize:       clientMetrics.SentMessageSize,
				SentMessageTime:       clientMetrics.MessageSendTime,
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
				Metadata: &message.Metadata{
					PayloadSize: 10,
					Creation:    time.Now().Add(-1 * time.Second),
					Type:        "goomerang.test.MessageV1",
				},
				Payload: &testMessages.MessageV1{},
			}

			m(h).Handle(sender, msg)

			AssertMetricsHandler(t, promhttp.Handler(), c.side)
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

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_inflight_duration_seconds_bucket{type="goomerang.test.MessageV1",le="2.5"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_inflight_duration_seconds_sum{type="goomerang.test.MessageV1"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_inflight_duration_seconds_count{type="goomerang.test.MessageV1"} 1`, system))

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_size_bytes_bucket{type="goomerang.test.MessageV1",le="10"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_size_bytes_sum{type="goomerang.test.MessageV1"} 10`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_received_message_size_bytes_count{type="goomerang.test.MessageV1"} 1`, system))

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_sent_message_size_bytes_bucket{type="goomerang.test.MessageV1",le="+Inf"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_sent_message_size_bytes_sum{type="goomerang.test.MessageV1"} 20`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_sent_message_size_bytes_count{type="goomerang.test.MessageV1"} 1`, system))

	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_send_duration_seconds_bucket{type="goomerang.test.MessageV1",le="+Inf"} 1`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_send_duration_seconds_sum{type="goomerang.test.MessageV1"}`, system))
	assert.Contains(t, body, fmt.Sprintf(`goomerang_%s_message_send_duration_seconds_count{type="goomerang.test.MessageV1"} 1`, system))
}

type fakeSender struct{}

func (f *fakeSender) Send(msg *message.Message) (payloadSize int, err error) {
	msg.Metadata = &message.Metadata{
		Type: "goomerang.test.MessageV1",
	}
	return 20, nil
}

func TestPromMiddlewareDoesValidations(t *testing.T) {
	_, err := middleware.PromHistograms(middleware.PromConfig{})
	require.EqualError(t, err, "metrics middleware config: validation error: validate: messageInflightTime be non nil")
}

func TestPromConfig_Validate_MessageInflightTime(t *testing.T) {
	c := middleware.PromConfig{
		ReceivedMessageSize:   clientMetrics.ReceivedMessageSize,
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
		SentMessageSize:       clientMetrics.SentMessageSize,
	}
	assert.EqualError(t, c.Validate(), "validate: messageInflightTime be non nil")
}

func TestPromConfig_Validate_NoReceivedMessageSize(t *testing.T) {
	c := middleware.PromConfig{
		MessageInflightTime:   clientMetrics.MessageInflightTime,
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
		SentMessageSize:       clientMetrics.SentMessageSize,
	}
	assert.EqualError(t, c.Validate(), "validate: receivedMessageSize be non nil")
}

func TestPromConfig_Validate_MessageProcessingTime(t *testing.T) {
	c := middleware.PromConfig{
		MessageInflightTime: clientMetrics.MessageInflightTime,
		ReceivedMessageSize: clientMetrics.ReceivedMessageSize,
		SentMessageSize:     clientMetrics.SentMessageSize,
	}
	assert.EqualError(t, c.Validate(), "validate: MessageProcessingTime be non nil")
}

func TestPromConfig_Validate_SentMessageSize(t *testing.T) {
	c := middleware.PromConfig{
		MessageInflightTime:   clientMetrics.MessageInflightTime,
		ReceivedMessageSize:   clientMetrics.ReceivedMessageSize,
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
	}
	assert.EqualError(t, c.Validate(), "validate: SentMessageSize be non nil")
}

func TestPromConfig_Validate_SentMessageTime(t *testing.T) {
	c := middleware.PromConfig{
		MessageInflightTime:   clientMetrics.MessageInflightTime,
		ReceivedMessageSize:   clientMetrics.ReceivedMessageSize,
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
		SentMessageSize:       clientMetrics.SentMessageSize,
	}
	assert.EqualError(t, c.Validate(), "validate: SentMessageTime be non nil")
}
