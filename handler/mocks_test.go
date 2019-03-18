package handler_test

import (
	"net/http/httptest"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/client"
	"github.com/davidsbond/sse-cluster/message"
	"github.com/stretchr/testify/mock"
)

type (
	MockBroker struct {
		mock.Mock

		clients map[string]*client.Client
	}

	ResponseRecorder struct {
		*httptest.ResponseRecorder

		close chan bool
	}
)

func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		close:            make(chan bool, 1),
	}
}

func (rr *ResponseRecorder) CloseNotify() <-chan bool {
	return rr.close
}

func (m *MockBroker) Status() *broker.Status {
	args := m.Called()

	if args.Get(0) != nil {
		return args.Get(0).(*broker.Status)
	}

	return nil
}

func (m *MockBroker) Publish(channel string, msg message.Message) {
	if cl, ok := m.clients[channel]; ok {
		cl.Write(msg.Bytes())
	}

	m.Called(channel, msg)
}

func (m *MockBroker) NewClient(channel string, clientID string) *client.Client {
	m.Called(channel, clientID)

	cl := client.New(clientID)
	m.clients[channel] = cl

	return cl
}

func (m *MockBroker) RemoveClient(channel string, client string) {
	delete(m.clients, channel)

	m.Called(channel, client)
}
