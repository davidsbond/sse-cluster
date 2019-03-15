package message

import (
	"encoding/json"
	"fmt"
)

type (
	// The Message type represents a server-sent event.
	Message struct {
		ID      string          `json:"id"`
		Event   string          `json:"event"`
		Data    json.RawMessage `json:"data"`
		Channel string          `json:"channel"`
		BeenTo  []string        `json:"been_to"`
	}
)

// Bytes returns the Message instance in its textual form
func (m *Message) Bytes() []byte {
	return []byte(fmt.Sprintf("id:%s\ndata:%s\nevent:%s\n\n", m.ID, m.Data, m.Event))
}

// JSON returns the Mesage instance in JSON encoding
func (m *Message) JSON() []byte {
	data, _ := json.Marshal(m)

	return data
}
