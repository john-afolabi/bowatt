package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

const internalErrorMessage = "the server could not complete the request"

type apiError struct {
	status  int
	code    string
	message string
}

func (e *apiError) Error() string { return e.code + ": " + e.message }

func badRequest(code, format string, args ...any) *apiError {
	return &apiError{status: http.StatusBadRequest, code: code, message: fmt.Sprintf(format, args...)}
}

func notFound(code, format string, args ...any) *apiError {
	return &apiError{status: http.StatusNotFound, code: code, message: fmt.Sprintf(format, args...)}
}

func tooLarge(code, format string, args ...any) *apiError {
	return &apiError{status: http.StatusRequestEntityTooLarge, code: code, message: fmt.Sprintf(format, args...)}
}

func unsupportedMedia(code, format string, args ...any) *apiError {
	return &apiError{status: http.StatusUnsupportedMediaType, code: code, message: fmt.Sprintf(format, args...)}
}

type errorEnvelope struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func errorBody(code, message string) errorEnvelope {
	return errorEnvelope{Error: errorDetail{Code: code, Message: message}}
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, context.Canceled) {
		return // client hung up; the access log records 499
	}

	var apiErr *apiError
	if errors.As(err, &apiErr) {
		s.writeJSON(w, apiErr.status, errorBody(apiErr.code, apiErr.message))
		return
	}

	s.log.ErrorContext(r.Context(), "request failed",
		slog.String("error", err.Error()),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("request_id", requestIDFrom(r.Context())),
	)
	s.writeJSON(w, http.StatusInternalServerError, errorBody("internal_error", internalErrorMessage))
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		s.log.Error("marshal response", slog.String("error", err.Error()))
		http.Error(w, `{"error":{"code":"internal_error","message":"response encoding failed"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		s.log.Debug("write response body", slog.String("error", err.Error()))
	}
}
