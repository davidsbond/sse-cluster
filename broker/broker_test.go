package broker_test

import (
	"net"
	"net/http"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/message"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBroker_Publish(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

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

				m.On("NumMembers").Return(2)

				m.On("Members").Return([]*memberlist.Node{
					{
						Name: "test",
						Addr: net.ParseIP("127.0.0.1"),
						Meta: []byte("8080"),
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

			b := broker.New(m, http.DefaultClient)
			c, _ := b.NewClient(tc.Channel, tc.Client)

			b.Publish(tc.Channel, "", tc.Message)

			result := <-c.Messages()

			assert.Equal(t, tc.Message.Bytes(), result)

			<-time.After(time.Millisecond * 250)
			assert.Equal(t, true, gock.IsDone())
		})
	}
}

func TestBroker_Status(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name                 string
		ExpectationFunc      func(*mock.Mock)
		ExpectedMemberCount  int
		ExpectedMembers      map[string]int
		ExpectedChannelCount int
	}{
		{
			Name:                 "It obtain node status",
			ExpectedChannelCount: 1,
			ExpectedMemberCount:  1,
			ExpectedMembers: map[string]int{
				"127.0.0.1": 1337,
			},
			ExpectationFunc: func(m *mock.Mock) {
				m.On("NumMembers").Return(1)
				m.On("LocalNode").Return(&memberlist.Node{
					Name: "test",
					Addr: net.ParseIP("127.0.0.1"),
					Port: 1337,
				})

				m.On("Members").Return([]*memberlist.Node{
					{
						Name: "test",
						Addr: net.ParseIP("127.0.0.1"),
						Port: 1337,
						Meta: []byte("8080"),
					},
				})
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := &MockMemberlist{}
			tc.ExpectationFunc(&m.Mock)

			b := broker.New(m, http.DefaultClient)
			b.NewClient("test", "test")

			status := b.Status()

			assert.NotNil(t, status)
			assert.Equal(t, tc.ExpectedMemberCount, status.Gossip.MemberCount)
			assert.Len(t, status.Channels, tc.ExpectedChannelCount)

			for addr, port := range tc.ExpectedMembers {
				act, ok := status.Gossip.Members[addr]

				assert.True(t, ok)
				assert.Equal(t, act, port)
			}
		})
	}

}

func TestBroker_NewClient(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name            string
		Channel         string
		Client          string
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

			b := broker.New(m, http.DefaultClient)
			cl, _ := b.NewClient(tc.Channel, tc.Client)

			assert.Equal(t, tc.Client, cl.ID())
		})
	}
}

func TestBroker_RemoveClient(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name            string
		Channel         string
		Client          string
		ExpectationFunc func(*mock.Mock)
	}{
		{
			Name:    "It should remove a client",
			Channel: "test",
			Client:  "test",
			ExpectationFunc: func(m *mock.Mock) {
				m.On("NumMembers").Return(1)
				m.On("LocalNode").Return(&memberlist.Node{
					Name: "test",
					Addr: net.ParseIP("127.0.0.1"),
					Port: 1337,
				})

				m.On("Members").Return([]*memberlist.Node{})
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := &MockMemberlist{}
			tc.ExpectationFunc(&m.Mock)

			b := broker.New(m, http.DefaultClient)
			b.NewClient(tc.Channel, tc.Client)

			status := b.Status()
			assert.Len(t, status.Channels, 1)

			b.RemoveClient(tc.Channel, tc.Client)

			status = b.Status()
			assert.Len(t, status.Channels, 0)
		})
	}
}
