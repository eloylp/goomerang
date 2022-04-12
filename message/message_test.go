package message_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/message"
)

func TestMessageJSON(t *testing.T) {
	msg := messageFixture()
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.JSONEq(t, `
{
 "header": {
   "k1": "v1",
   "k2": "v2"
 },
  "metadata": {
    "uuid": "09AF",
    "creation": "1970-01-01T00:00:00.001Z",
    "isSync": true,
    "payloadSize": 10,
    "type": "goomerang.test.MessageV1"
 },
  "payload": {
  "message": "Hi !"
 }
}
`, string(data))

}

func messageFixture() *message.Message {
	return &message.Message{
		Metadata: &message.Metadata{
			Creation:    time.UnixMilli(1).UTC(),
			UUID:        "09AF",
			Type:        "goomerang.test.MessageV1",
			PayloadSize: 10,
			IsSync:      true,
		},
		Payload: &test.MessageV1{Message: "Hi !"},
		Header: message.Header{
			"k1": "v1",
			"k2": "v2",
		},
	}
}

func TestMessageText(t *testing.T) {
	msg := messageFixture()
	possibleOutput1 := `metadata: creation=1970-01-01 00:00:00.001 +0000 UTC,uuid=09AF,type=goomerang.test.MessageV1,payloadSize=10,isSync=true headers: k1=v1,k2=v2 - payload: message:"Hi !"`
	possibleOutput2 := `metadata: creation=1970-01-01 00:00:00.001 +0000 UTC,uuid=09AF,type=goomerang.test.MessageV1,payloadSize=10,isSync=true headers: k2=v2,k1=v1 - payload: message:"Hi !"`
	textMsg := msg.String()
	if textMsg == possibleOutput1 {
		return
	}
	if textMsg == possibleOutput2 {
		return
	}
	t.Errorf("expected header serialization to at least accomplish one of the possible outputs (see test). Was %s", textMsg)
}
