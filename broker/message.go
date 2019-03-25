package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type (
	// The Message type represents a server-sent event.
	Message struct {
		// The event ID to set the EventSource object's last event ID value.
		ID string `json:"id"`

		// A string identifying the type of event described. If this is specified, an event will
		// be dispatched on the browser to the listener for the specified event name;
		// the website source code should use addEventListener() to listen for named events.
		// The onmessage handler is called if no event name is specified for a message.
		Event string `json:"event"`

		// The data field for the message. When the EventSource receives multiple consecutive lines that begin with data:,
		// it will concatenate them, inserting a newline character between each one.
		// Trailing newlines are removed.
		Data json.RawMessage `json:"data"`

		// The reconnection time to use when attempting to send the event. This must be an integer,
		// specifying the reconnection time in milliseconds.
		// If a non-integer value is specified, the field is ignored.
		Retry int `json:"retry"`

		// Contains identifiers of previous nodes this event has been through
		BeenTo []string `json:"been_to"`
	}
)

// Bytes returns the Message instance in its textual form
func (m *Message) Bytes() []byte {
	var out bytes.Buffer

	if m.ID != "" {
		out.WriteString("id: ")
		out.WriteString(m.ID)
		out.WriteRune('\n')
	}

	if m.Event != "" {
		out.WriteString("event: ")
		out.WriteString(m.Event)
		out.WriteRune('\n')
	}

	if m.Data != nil {
		out.WriteString("data: ")
		out.Write(m.Data)
		out.WriteRune('\n')
	}

	if m.Retry > 0 {
		out.WriteString("retry: ")
		out.WriteString(fmt.Sprint(m.Retry))
		out.WriteRune('\n')
	}

	out.WriteRune('\n')

	return out.Bytes()
}

// JSON returns the Mesage instance in JSON encoding
func (m *Message) JSON() []byte {
	data, _ := json.Marshal(m)

	return data
}
