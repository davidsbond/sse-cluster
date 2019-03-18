package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/handler"
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
