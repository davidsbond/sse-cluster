package message

import (
	"encoding/json"
	"fmt"
)

type (
	// The Message type represents a server-sent event.
	Message struct {
		ID     string          `json:"id" validate:"required"`
		Event  string          `json:"event" validate:"required"`
		Data   json.RawMessage `json:"data" validate:"required"`
		BeenTo []string        `json:"been_to,omitempty"`
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
