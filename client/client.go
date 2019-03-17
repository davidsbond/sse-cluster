package client

type (
	// The Client type represents a single client connected to the
	// broker
	Client struct {
		id   string
		data chan []byte
	}
)

// New creates a new instance of the Client type with the given
// identifier.
func New(id string) *Client {
	return &Client{
		id:   id,
		data: make(chan []byte, 1),
	}
}

// ID returns this client's identifier.
func (c *Client) ID() string {
	return c.id
}

// Write writes a given array of bytes to a client
func (c *Client) Write(data []byte) {
	c.data <- data
}

// Messages returns a read-only channel for this client's messages.
func (c *Client) Messages() <-chan []byte {
	return c.data
}
