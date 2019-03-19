package channel

import (
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

// Write writes a given message to a random client in the channel
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
func (c *Channel) AddClient(id string) *client.Client {
	c.mux.Lock()
	defer c.mux.Unlock()

	cl := client.New(id)
	c.clients[id] = cl

	return cl
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
