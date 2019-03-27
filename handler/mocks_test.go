package handler_test

import (
	"github.com/davidsbond/sse-cluster/broker"
	"github.com/stretchr/testify/mock"
)

type (
	MockBroker struct {
		mock.Mock

		clients map[string]*broker.Client
	}
)

func (m *MockBroker) Status() *broker.Status {
	args := m.Called()

	if args.Get(0) != nil {
		return args.Get(0).(*broker.Status)
	}

	return nil
}

func (m *MockBroker) Publish(channel, client string, msg broker.Message) error {
	if cl, ok := m.clients[channel]; ok {
		cl.Write(msg)
	}

	args := m.Called(channel, client, msg)

	return args.Error(0)
}

func (m *MockBroker) NewClient(channel string, clientID string) (*broker.Client, error) {
	args := m.Called(channel, clientID)

	cl := broker.NewClient(clientID)
	m.clients[channel] = cl

	return cl, args.Error(1)
}

func (m *MockBroker) RemoveClient(channel string, client string) {
	delete(m.clients, channel)

	m.Called(channel, client)
}
