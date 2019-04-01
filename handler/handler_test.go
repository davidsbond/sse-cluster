package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/handler"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Status(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

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
				m.On("Status").Return(&broker.Status{})
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
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name            string
		Channel         string
		Message         broker.Message
		ExpectedCode    int
		ExpectationFunc func(*mock.Mock)
	}{
		{
			Name:    "When message is published successfully, returns a 200",
			Channel: "success",
			Message: broker.Message{
				ID:    "test",
				Event: "test",
				Data:  []byte("{}"),
			},
			ExpectedCode: http.StatusOK,
			ExpectationFunc: func(m *mock.Mock) {
				m.On("Publish", "success", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := &MockBroker{clients: make(map[string]*broker.Client)}
			h := handler.New(m)

			tc.ExpectationFunc(&m.Mock)

			body, _ := json.Marshal(tc.Message)
			r := httptest.NewRequest("POST", "/publish/"+tc.Channel, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/publish/{channel}", h.Publish)

			router.ServeHTTP(w, r)

			assert.Equal(t, tc.ExpectedCode, w.Code)
		})
	}
}

func TestHandler_Subscribe(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.PanicLevel)

	tt := []struct {
		Name            string
		Channel         string
		Message         broker.Message
		ExpectedCode    int
		ExpectationFunc func(*mock.Mock)
	}{
		{
			Name:         "When subscription is successful, writes a 200",
			Channel:      "success",
			ExpectedCode: http.StatusOK,
			Message: broker.Message{
				ID:    "test",
				Event: "test",
				Data:  []byte("{}"),
			},
			ExpectationFunc: func(m *mock.Mock) {
				m.On("NewClient", "success", mock.Anything).Return(nil, nil)
				m.On("Publish", "success", mock.Anything, mock.Anything).Return(nil)
				m.On("RemoveClient", "success", mock.Anything).Return(nil)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := &MockBroker{clients: make(map[string]*broker.Client)}
			h := handler.New(m)

			tc.ExpectationFunc(&m.Mock)

			r := httptest.NewRequest("GET", "/subscribe/"+tc.Channel, nil)
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/subscribe/{channel}", h.Subscribe)

			ctx, cancel := context.WithCancel(r.Context())
			go router.ServeHTTP(w, r.WithContext(ctx))

			<-time.After(time.Millisecond * 100)

			if err := m.Publish(tc.Channel, "", tc.Message); err != nil {
				assert.Fail(t, err.Error())
			}

			<-time.After(time.Millisecond * 100)
			cancel()
			<-time.After(time.Millisecond * 100)

			assert.Equal(t, tc.ExpectedCode, w.Code)
			assert.Equal(t, tc.Message.Bytes(), w.Body.Bytes())
		})
	}
}
