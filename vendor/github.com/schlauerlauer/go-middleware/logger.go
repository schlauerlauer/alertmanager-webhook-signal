package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)

		if wrapped.statusCode < http.StatusBadRequest {
			slog.Info("request", "status", wrapped.statusCode, "method", r.Method, "path", r.URL, "forwarded", r.Header.Get("X-Forwarded-For"), "duration", time.Since(start))
		} else if wrapped.statusCode < http.StatusInternalServerError {
			slog.Warn("request", "status", wrapped.statusCode, "method", r.Method, "path", r.URL, "forwarded", r.Header.Get("X-Forwarded-For"), "duration", time.Since(start))
		} else {
			slog.Error("request", "status", wrapped.statusCode, "method", r.Method, "path", r.URL, "forwarded", r.Header.Get("X-Forwarded-For"), "duration", time.Since(start))
		}
	})
}

// doesn't wrap the response writer, therefore no status code is available in the logs, but w.Flusher will work
func LoggingDefault(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		slog.Info("request", "method", r.Method, "forwarded", r.Header.Get("X-Forwarded-For"), "path", r.URL, "duration", time.Since(start))
	})
}
