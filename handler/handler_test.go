package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/handler"
	"github.com/davidsbond/sse-cluster/message"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Status(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name                string
		ExpectedCode        int
		ExpectedContentType string
		ExpectationFunc     func(*mock.Mock)
	}{
		{
			Name:                "It should get node status",
			ExpectedCode:        http.StatusOK,
			ExpectedContentType: "application/json",
			ExpectationFunc: func(m *mock.Mock) {
				m.On("GetStatus").Return(&broker.Status{})
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := &MockBroker{}
			tc.ExpectationFunc(&m.Mock)

			h := handler.New(m)
			r := httptest.NewRequest("GET", "/status", nil)
			w := httptest.NewRecorder()

			h.Status(w, r)

			assert.Equal(t, tc.ExpectedCode, w.Code)
			assert.Equal(t, tc.ExpectedContentType, w.Header().Get("Content-Type"))
		})
	}
}

func TestHandler_Publish(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name            string
		Channel         string
		Message         message.Message
		ExpectedCode    int
		ExpectationFunc func(*mock.Mock)
	}{
		{
			Name:    "When message is published successfully, returns a 200",
			Channel: "success",
			Message: message.Message{
				ID:    "test",
				Event: "test",
				Data:  []byte("{}"),
			},
			ExpectedCode: http.StatusOK,
			ExpectationFunc: func(m *mock.Mock) {
				m.On("Publish", "success", mock.Anything).Return(nil)
			},
		},
		{
			Name:    "When message is published unsuccessfully, returns a 500",
			Channel: "error",
			Message: message.Message{
				ID:    "test",
				Event: "test",
				Data:  []byte("{}"),
			},
			ExpectedCode: http.StatusInternalServerError,
			ExpectationFunc: func(m *mock.Mock) {
				m.On("Publish", "error", mock.Anything).Return(errors.New("test"))
			},
		},
		{
			Name:            "When message fails validation, returns a 400",
			Channel:         "channel",
			Message:         message.Message{},
			ExpectedCode:    http.StatusBadRequest,
			ExpectationFunc: func(m *mock.Mock) {},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := &MockBroker{}
			h := handler.New(m)

			tc.ExpectationFunc(&m.Mock)

			body, _ := json.Marshal(tc.Message)
			r := httptest.NewRequest("POST", "/publish/"+tc.Channel, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			mux := mux.NewRouter()
			mux.HandleFunc("/publish/{channel}", h.Publish)

			mux.ServeHTTP(w, r)

			assert.Equal(t, tc.ExpectedCode, w.Code)
		})
	}
}
