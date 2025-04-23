package middleware

import (
	"log/slog"
	"net/http"
)

func NewLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("incoming request",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
			)
			next.ServeHTTP(w, r)
		})
	}
}
