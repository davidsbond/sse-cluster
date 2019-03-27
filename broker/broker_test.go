package broker_test

import (
	"net"
	"net/http"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	"github.com/davidsbond/sse-cluster/broker"
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
		Message         broker.Message
		ExpectationFunc func(*mock.Mock, *gock.Request)
	}{
		{
			Name:    "It should write a message to the client",
			Channel: "test",
			Client:  "test",
			Message: broker.Message{
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
		{
			Name:    "It should write a message to the channel",
			Channel: "test",
			Message: broker.Message{
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
		{
			Name: "It should write a message to all channels",
			Message: broker.Message{
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
			defer b.Close()

			c, err := b.NewClient(tc.Channel, tc.Client)

			if err != nil {
				assert.Fail(t, err.Error())
				return
			}

			if err := b.Publish(tc.Channel, tc.Client, tc.Message); err != nil {
				assert.Fail(t, err.Error())
				return
			}

			result := <-c.Messages()

			assert.Equal(t, tc.Message, result)

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
			defer b.Close()

			if _, err := b.NewClient("test", "test"); err != nil {
				assert.Fail(t, err.Error())
				return
			}

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
			defer b.Close()

			cl, err := b.NewClient(tc.Channel, tc.Client)

			if err != nil {
				assert.Fail(t, err.Error())
				return
			}

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
			defer b.Close()

			if _, err := b.NewClient(tc.Channel, tc.Client); err != nil {
				assert.Fail(t, err.Error())
				return
			}

			status := b.Status()
			assert.Len(t, status.Channels, 1)

			b.RemoveClient(tc.Channel, tc.Client)

			status = b.Status()
			assert.Len(t, status.Channels, 0)
		})
	}
}
