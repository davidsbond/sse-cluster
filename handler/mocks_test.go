package handler_test

import (
	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/client"
	"github.com/davidsbond/sse-cluster/message"
	"github.com/stretchr/testify/mock"
)

type (
	MockBroker struct {
		mock.Mock
	}
)

func (m *MockBroker) GetStatus() *broker.Status {
	args := m.Called()

	if args.Get(0) != nil {
		return args.Get(0).(*broker.Status)
	}

	return nil
}

func (m *MockBroker) Publish(channel string, msg message.Message) error {
	return m.Called(channel, msg).Error(0)
}

func (m *MockBroker) NewClient(channel string, clientID string) *client.Client {
	args := m.Called(channel, clientID)

	if args.Get(0) != nil {
		return args.Get(0).(*client.Client)
	}

	return nil
}

func (m *MockBroker) RemoveClient(channel string, client string) {
	m.Called(channel, client)
}
