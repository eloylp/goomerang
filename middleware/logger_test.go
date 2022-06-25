package middleware_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/messaging"
	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/middleware"
)

func TestLogger(t *testing.T) {
	output := bytes.NewBuffer(nil)

	logger := logrus.New()
	logger.SetOutput(output)
	logger.SetLevel(logrus.DebugLevel)

	nullHandler := message.HandlerFunc(func(s message.Sender, msg *message.Message) {})
	date, err := time.Parse(time.RFC3339, "2022-04-10T22:56:18Z")
	require.NoError(t, err)

	payload := &testMessages.MessageV1{Message: "hi!"}
	msg := &message.Message{
		Metadata: message.Metadata{
			Creation:    date,
			UUID:        "09AF",
			Kind:        messaging.FQDN(payload),
			PayloadSize: 10,
			IsSync:      false,
		},
		Payload: payload,
		Header: map[string]string{
			"k1": "v1",
			"k2": "v2",
		},
	}

	m := middleware.Logger(logger)
	m(nullHandler).Handle(&fakeSender{}, msg)
	logLines := output.String()
	assert.Contains(t, logLines, `msg="message processed"`)
	AssertHeadersArePresent(t, logLines)
	assert.Contains(t, logLines, `metadata="creation=2022-04-10 22:56:18 +0000 UTC,uuid=09AF,kind=goomerang.test.MessageV1,payloadSize=10,isSync=false"`)
	assert.Contains(t, logLines, `payload="message:\"hi!\""`)
	assert.Contains(t, logLines, `processingTime=`)
	assert.Contains(t, logLines, `kind=goomerang.test.MessageV1`)
}

func AssertHeadersArePresent(t *testing.T, lines string) {
	if strings.Contains(lines, `headers="k1=v1,k2=v2"`) {
		return
	}
	if strings.Contains(lines, `headers="k2=v2,k1=v1"`) {
		return
	}
	t.Errorf("headers should match at least one example. Got %s", lines)
}
