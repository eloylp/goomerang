package messaging

import (
	"fmt"
	"strings"

	"go.eloylp.dev/goomerang/message"
)

const userMessageHeadersPrefix = "message-"

func packHeaders(headers message.Header) message.Header {
	packedHeaders := message.Header{}
	for k, v := range headers {
		packedHeaders.Set(fmt.Sprintf("%s%s", userMessageHeadersPrefix, k), v)
	}
	return packedHeaders
}

func unpackHeaders(headers message.Header) message.Header {
	unpackedHeaders := message.Header{}
	for k, v := range headers {
		if strings.HasPrefix(k, userMessageHeadersPrefix) {
			key := strings.TrimPrefix(k, userMessageHeadersPrefix)
			unpackedHeaders.Set(key, v)
		}
	}
	return unpackedHeaders
}
