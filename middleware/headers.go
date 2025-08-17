package middleware

import (
	"net/http"
)

func ImageResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Thumbnailer", "Thumbra")
		w.Header().Set("Cache-Control", "public, max-age=31536000")

		next.ServeHTTP(w, r)
	})
}
