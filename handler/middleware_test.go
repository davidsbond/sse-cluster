package handler_test

import (
	"github.com/davidsbond/sse-cluster/handler"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_CORS(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name            string
		Method          string
		ExpectedHeaders map[string]string
		ExpectedStatus  int
	}{
		{
			Name:   "It should add CORS headers to general requests",
			Method: "GET",
			ExpectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "It should add CORS headers to OPTIONS requests",
			Method: "OPTIONS",
			ExpectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Authorization",
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			r := httptest.NewRequest(tc.Method, "/", nil)
			w := httptest.NewRecorder()

			mux := mux.NewRouter()

			mux.Use(handler.CORSMiddleware)
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
			mux.ServeHTTP(w, r)

			assert.Equal(t, tc.ExpectedStatus, w.Code)

			for expKey, expVal := range tc.ExpectedHeaders {
				actVal := w.Header().Get(expKey)

				assert.Equal(t, expVal, actVal)
			}
		})
	}
}
