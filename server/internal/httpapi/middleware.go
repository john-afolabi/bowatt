package httpapi

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/oklog/ulid/v2"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

// The client hung up first.
const statusClientClosedRequest = 499

func requestIDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDContextKey).(string); ok {
		return id
	}
	return ""
}

func (s *Server) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := ulid.Make().String()
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDContextKey, id)))
	})
}

func (s *Server) withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		switch {
		case status == 0 && r.Context().Err() != nil:
			status = statusClientClosedRequest
		case status == 0:
			status = http.StatusOK
		}

		level := slog.LevelInfo
		switch {
		case r.URL.Path == "/healthz", status == statusClientClosedRequest:
			level = slog.LevelDebug
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case status >= http.StatusBadRequest:
			level = slog.LevelWarn
		}

		s.log.Log(r.Context(), level, "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Int("bytes", recorder.bytes),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			slog.String("remote", r.RemoteAddr),
			slog.String("request_id", requestIDFrom(r.Context())),
		)
	})
}

func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w}

		defer func() {
			if recovered := recover(); recovered != nil {
				s.log.ErrorContext(r.Context(), "panic recovered",
					slog.Any("panic", recovered),
					slog.String("path", r.URL.Path),
					slog.String("request_id", requestIDFrom(r.Context())),
					slog.String("stack", string(debug.Stack())),
				)
				if !recorder.wroteHeader {
					s.writeJSON(recorder, http.StatusInternalServerError, errorBody("internal_error", internalErrorMessage))
				}
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(p)
	r.bytes += n
	return n, err
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer %T cannot hijack", r.ResponseWriter)
	}
	r.wroteHeader = true
	r.status = http.StatusSwitchingProtocols
	return hijacker.Hijack()
}
