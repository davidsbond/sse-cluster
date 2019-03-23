package channel

import (
	"fmt"
	"sync"

	"github.com/davidsbond/sse-cluster/client"
	"github.com/sirupsen/logrus"
)

type (
	// The Channel type represents a channel within the broker. Each channel
	// has a unique identifier and can have one or more clients. When events
	// are published to a channel, a client is chosen at random
	Channel struct {
		id      string
		clients map[string]*client.Client
		mux     sync.Mutex
		log     *logrus.Entry
	}
)

// New creates a new instance of the Channel type using the given identifier
func New(id string) *Channel {
	return &Channel{
		id:      id,
		clients: make(map[string]*client.Client),
		log:     logrus.WithField("channel", id),
	}
}

// WriteTo writes a message directly to a given client
func (c *Channel) WriteTo(clientID string, msg []byte) {
	c.log.WithField("clientId", clientID).Info("writing message to client")

	c.mux.Lock()
	defer c.mux.Unlock()

	if cl, ok := c.clients[clientID]; ok {
		cl.Write(msg)
	}
}

// Write writes a given message to all clients in the channel
func (c *Channel) Write(msg []byte) {
	c.log.Info("writing message to channel")

	for _, cl := range c.clients {
		cl.Write(msg)

		c.log.WithField("client", cl.ID()).Info("writing message to client")
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

// AddClient adds a new client to the channel
func (c *Channel) AddClient(id string) (*client.Client, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.clients[id]; ok {
		return nil, fmt.Errorf("failed to add client to channel %s, client with id %s already exists", c.id, id)
	}

	cl := client.New(id)
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
