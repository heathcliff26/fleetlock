package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (res *responseWrapper) WriteHeader(statusCode int) {
	res.ResponseWriter.WriteHeader(statusCode)
	res.statusCode = statusCode
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		wrapped := &responseWrapper{
			ResponseWriter: res,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, req)

		slog.Debug("Got Request",
			slog.String("source", ReadUserIP(req)),
			slog.Int("status", wrapped.statusCode),
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.Any("took", time.Since(start)),
		)
	})
}

func ReadUserIP(req *http.Request) string {
	IPAddress := req.Header.Get("x-real-ip")
	if IPAddress == "" {
		IPAddress = req.Header.Get("x-forwarded-for")
	}
	if IPAddress == "" {
		IPAddress = req.RemoteAddr
	}
	return IPAddress
}
