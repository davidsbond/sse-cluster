package client

import (
	"github.com/davidsbond/sse-cluster/message"
)

type (
	// The Client type represents a single client connected to the
	// broker
	Client struct {
		id       string
		messages chan message.Message
	}
)

// New creates a new instance of the Client type with the given
// identifier.
func New(id string) *Client {
	return &Client{
		id:       id,
		messages: make(chan message.Message, 1),
	}
}

// ID returns this client's identifier.
func (c *Client) ID() string {
	return c.id
}

// Write writes a given array of bytes to a client
func (c *Client) Write(msg message.Message) {
	c.messages <- msg
}

// Messages returns a read-only channel for this client's messages.
func (c *Client) Messages() <-chan message.Message {
	return c.messages
}
