package channel

import (
	"math/rand"
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
	}
)

// New creates a new instance of the Channel type using the given identifier
func New(id string) *Channel {
	return &Channel{
		id:      id,
		clients: make(map[string]*client.Client),
	}
}

// Write writes a given message to a random client in the channel
func (c *Channel) Write(msg []byte) {
	logrus.WithField("channel", c.id).Info("writing message to channel")

	client := c.randomClient()
	client.Write(msg)

	logrus.WithField("client", client.ID()).Info("writing message to client")
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

// randomClient returns a random client in the map of clients for
// the channel.
func (c *Channel) randomClient() *client.Client {
	c.mux.Lock()
	defer c.mux.Unlock()

	i := rand.Intn(len(c.clients))

	var k string
	for k = range c.clients {
		if i == 0 {
			break
		}

		i--
	}

	return c.clients[k]
}
