package channel_test

import (
	"testing"

	"github.com/davidsbond/sse-cluster/channel"
	"github.com/davidsbond/sse-cluster/message"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestChannel_Write(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name    string
		Client  string
		Channel string
		Message message.Message
	}{
		{
			Name:    "It should write a message",
			Client:  "test",
			Channel: "test",
			Message: message.Message{
				Data: []byte("test"),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			ch := channel.New(tc.Channel)
			cl, _ := ch.AddClient(tc.Client)

			ch.Write(tc.Message)
			msg := <-cl.Messages()

			assert.Equal(t, tc.Message, msg)
		})
	}
}

func TestChannel_AddClient(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name    string
		Client  string
		Channel string
	}{
		{
			Name:    "It should create a client",
			Client:  "test",
			Channel: "test",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			ch := channel.New(tc.Channel)
			cl, _ := ch.AddClient(tc.Client)

			assert.NotNil(t, ch)
			assert.Equal(t, tc.Client, cl.ID())
		})
	}
}

func TestChannel_ClientIDs(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name    string
		Client  string
		Channel string
	}{
		{
			Name:    "It should list client identifiers",
			Client:  "test",
			Channel: "test",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			ch := channel.New(tc.Channel)
			cl, _ := ch.AddClient(tc.Client)

			assert.NotNil(t, ch)
			assert.Contains(t, ch.ClientIDs(), cl.ID())
		})
	}
}

func TestChannel_NumClients(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name                string
		Client              string
		Channel             string
		ExpectedClientCount int
	}{
		{
			Name:                "It should return the number of clients",
			Client:              "test",
			Channel:             "test",
			ExpectedClientCount: 1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			ch := channel.New(tc.Channel)
			ch.AddClient(tc.Client)

			assert.NotNil(t, ch)
			assert.Equal(t, tc.ExpectedClientCount, ch.NumClients())
		})
	}
}

func TestChannel_RemoveClient(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name    string
		Client  string
		Channel string
	}{
		{
			Name:    "It should create a client",
			Client:  "test",
			Channel: "test",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			ch := channel.New(tc.Channel)
			cl, _ := ch.AddClient(tc.Client)

			assert.NotNil(t, ch)
			assert.Equal(t, tc.Client, cl.ID())

			ch.RemoveClient(cl.ID())
			assert.Equal(t, 0, ch.NumClients())
		})
	}
}
