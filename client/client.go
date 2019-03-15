package client

type (
	// The Client type represents a single client connected to the
	// broker
	Client struct {
		id   string
		data chan []byte
	}
)

func New(id string) *Client {
	return &Client{
		id:   id,
		data: make(chan []byte, 1),
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Write(data []byte) {
	c.data <- data
}

func (c *Client) Messages() <-chan []byte {
	return c.data
}
