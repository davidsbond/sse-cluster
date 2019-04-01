package broker

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type (
	// The Channel type represents a channel within the broker. Each channel
	// has a unique identifier and can have one or more clients. When events
	// are published to a channel, a client is chosen at random
	Channel struct {
		id      string
		clients map[string]*Client
		mux     sync.Mutex
		log     *logrus.Entry
	}
)

// NewChannel creates a new instance of the Channel type using the given identifier
func NewChannel(id string) *Channel {
	return &Channel{
		id:      id,
		clients: make(map[string]*Client),
		log:     logrus.WithField("channel", id),
	}
}

// WriteTo writes a message directly to a given client
func (c *Channel) WriteTo(clientID string, msg Message) {
	c.log.WithFields(logrus.Fields{
		"clientId": clientID,
		"eventId":  msg.ID,
		"event":    msg.Event,
	}).Info("writing message to client")

	c.mux.Lock()
	defer c.mux.Unlock()

	if cl, ok := c.clients[clientID]; ok {
		cl.Write(msg)

		c.log.WithFields(logrus.Fields{
			"client":  cl.ID(),
			"eventId": msg.ID,
			"event":   msg.Event,
		}).Info("wrote message to client")

		return nil
	}
}

// Write writes a given message to all clients in the channel
func (c *Channel) Write(msg Message) {
	c.log.WithFields(logrus.Fields{
		"eventId": msg.ID,
		"event":   msg.Event,
	}).Info("writing message to channel")

	c.mux.Lock()
	defer c.mux.Unlock()

	for _, cl := range c.clients {
		c.mux.Unlock()

		cl.Write(msg)

		c.mux.Lock()

		c.log.WithFields(logrus.Fields{
			"client":  cl.ID(),
			"eventId": msg.ID,
			"event":   msg.Event,
		}).Info("wrote message to client")
	}
}

// ClientIDs returns an array of all client identifiers in this
// channel.
func (c *Channel) ClientIDs() []string {
	var out []string

	c.mux.Lock()
	defer c.mux.Unlock()

	for id := range c.clients {
		out = append(out, id)
	}

	return out
}

// NewClient adds a new client to the channel
func (c *Channel) NewClient(id string) (*Client, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.clients[id]; ok {
		return nil, fmt.Errorf("failed to add client to channel %s, client with id %s already exists", c.id, id)
	}

	cl := NewClient(id)
	c.clients[id] = cl

	return cl, nil
}

// NumClients returns the total number of clients for a
// channel.
func (c *Channel) NumClients() int {
	c.mux.Lock()
	defer c.mux.Unlock()

	return len(c.clients)
}

// RemoveClient removes a client from the channel
func (c *Channel) RemoveClient(id string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	delete(c.clients, id)
}
