package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		if lrw.statusCode > 400 {
			slog.ErrorContext(r.Context(), "HTTP Request failed",
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", lrw.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			slog.InfoContext(r.Context(), "HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", lrw.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		}
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	if !lrw.written {
		lrw.statusCode = code
		lrw.written = true
	}
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	if !lrw.written {
		lrw.written = true
	}
	return lrw.ResponseWriter.Write(data)
}
