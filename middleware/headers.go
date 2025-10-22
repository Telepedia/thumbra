package middleware

import (
	"net/http"
)

func ImageResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "Thumbra")

		next.ServeHTTP(w, r)
	})
}
