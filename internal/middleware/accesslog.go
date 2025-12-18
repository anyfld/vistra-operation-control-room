package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(b)
	r.size += n

	return n, err
}

func Middleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			start := time.Now()

			logger.Info(
				"access",
				"event", "inbound",
				"method", req.Method,
				"path", req.URL.Path,
				"remote_ip", req.RemoteAddr,
			)

			rec := &statusRecorder{ResponseWriter: writer, status: 0, size: 0}

			next.ServeHTTP(rec, req)

			if rec.status == 0 {
				rec.status = http.StatusOK
			}

			logger.Info(
				"access",
				"event", "outbound",
				"method", req.Method,
				"path", req.URL.Path,
				"remote_ip", req.RemoteAddr,
				"status", rec.status,
				"size", rec.size,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
