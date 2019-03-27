package broker_test

import (
	"github.com/davidsbond/sse-cluster/broker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage_Bytes(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name          string
		Message       broker.Message
		ExpectedBytes []byte
	}{
		{
			Name: "It should include id",
			Message: broker.Message{
				ID: "123",
			},
			ExpectedBytes: []byte("id: 123\n\n"),
		},
		{
			Name: "It should include event",
			Message: broker.Message{
				Event: "test",
			},
			ExpectedBytes: []byte("event: test\n\n"),
		},
		{
			Name: "It should include retry",
			Message: broker.Message{
				Retry: 100,
			},
			ExpectedBytes: []byte("retry: 100\n\n"),
		},
		{
			Name: "It should include data",
			Message: broker.Message{
				Data: []byte("{}"),
			},
			ExpectedBytes: []byte("data: {}\n\n"),
		},
		{
			Name: "It should include everything",
			Message: broker.Message{
				ID:    "123",
				Event: "test",
				Retry: 100,
				Data:  []byte("{}"),
			},
			ExpectedBytes: []byte("id: 123\nevent: test\ndata: {}\nretry: 100\n\n"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			bytes := tc.Message.Bytes()

			assert.Equal(t, tc.ExpectedBytes, bytes)
		})
	}
}
