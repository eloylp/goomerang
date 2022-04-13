package message

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
)

type Message struct {
	Metadata *Metadata     `json:"metadata"`
	Payload  proto.Message `json:"payload"`
	Header   Header        `json:"header"`
}

func New() *Message {
	return &Message{
		Header: Header{},
	}
}

func (m *Message) SetPayload(p proto.Message) *Message {
	m.Payload = p
	return m
}

func (m *Message) SetHeader(k, v string) *Message {
	m.Header.Set(k, v)
	return m
}

func (m *Message) GetHeader(k string) string {
	return m.Header.Get(k)
}

func (m Message) String() string {
	return fmt.Sprintf("metadata: %s headers: %s - payload: %s", m.Metadata, m.Header, m.Payload)
}

type Metadata struct {
	Creation    time.Time `json:"creation"`
	UUID        string    `json:"uuid"`
	Type        string    `json:"type"`
	PayloadSize int       `json:"payloadSize"`
	IsSync      bool      `json:"isSync"`
}

func (m Metadata) String() string {
	elems := make([]string, 0, 5)
	elems = append(elems, fmt.Sprintf("creation=%v", m.Creation))
	elems = append(elems, fmt.Sprintf("uuid=%s", m.UUID))
	elems = append(elems, fmt.Sprintf("type=%s", m.Type))
	elems = append(elems, fmt.Sprintf("payloadSize=%v", m.PayloadSize))
	elems = append(elems, fmt.Sprintf("isSync=%v", m.IsSync))
	return strings.Join(elems, ",")
}

type Header map[string]string

func (h Header) String() string {
	var elems []string
	for k, v := range h {
		elems = append(elems, fmt.Sprintf("%v=%v", k, v))
	}
	return strings.Join(elems, ",")
}

func (h Header) Get(key string) string {
	return h[key]
}

func (h Header) Set(key, value string) {
	h[key] = value
}
