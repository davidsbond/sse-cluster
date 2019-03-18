package broker_test

import (
	"net"
	"testing"

	"gopkg.in/h2non/gock.v1"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/message"
	"github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBroker_Publish(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name            string
		Channel         string
		Client          string
		Message         message.Message
		ExpectationFunc func(*mock.Mock, *gock.Request)
	}{
		{
			Name:    "It should write a message to the client",
			Channel: "test",
			Client:  "test",
			Message: message.Message{
				ID:    "test",
				Event: "test",
				Data:  []byte("{}"),
			},
			ExpectationFunc: func(m *mock.Mock, g *gock.Request) {
				m.On("LocalNode").Return(&memberlist.Node{
					Name: "test",
				})

				m.On("Members").Return([]*memberlist.Node{
					{
						Name: "test",
						Addr: net.ParseIP("127.0.0.1"),
					},
				})

				g.Post("/publish").Reply(200)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			defer gock.Off()
			req := gock.New("http://127.0.0.1:8080")

			m := &MockMemberlist{}
			tc.ExpectationFunc(&m.Mock, req)

			b := broker.New(m, "8080")
			c := b.NewClient(tc.Channel, tc.Client)

			b.Publish(tc.Channel, tc.Message)

			result := <-c.Messages()

			assert.Equal(t, tc.Message.Bytes(), result)
			assert.Equal(t, true, gock.IsDone())
		})
	}
}

func TestBroker_Status(t *testing.T) {
	t.Parallel()

}

func TestBroker_NewClient(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name    string
		Channel string
		Client  string
		ExpectationFunc func(*mock.Mock)
	}{
		{
			Name:    "It should create a new client",
			Channel: "test",
			Client:  "test",
			ExpectationFunc: func(m *mock.Mock) {
				m.On("LocalNode").Return(&memberlist.Node{
					Name: "test",
				})
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := &MockMemberlist{}
			tc.ExpectationFunc(&m.Mock)
		
			b := broker.New(m, "")
			cl := b.NewClient(tc.Channel, tc.Client)

			assert.Equal(t, tc.Client, cl.ID())
		})
	}
}

func TestBroker_RemoveClient(t *testing.T) {
	t.Parallel()

}
