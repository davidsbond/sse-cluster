package handler

import "net/http"

// CORSMiddleware is an HTTP middleware that adds cross-origin headers to the
// HTTP response writer.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		}

		next.ServeHTTP(w, r)
	})
}
