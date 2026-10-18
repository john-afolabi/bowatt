package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"bowatt/internal/config"
	"bowatt/internal/store"
)

type Server struct {
	cfg   config.Config
	log   *slog.Logger
	store *store.SQLite
}

func New(cfg config.Config, log *slog.Logger, st *store.SQLite) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{cfg: cfg, log: log, store: st}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/images", s.handleListImages)
	mux.HandleFunc("POST /uploads", s.handleCreateUpload)
	mux.HandleFunc("GET /uploads/{key}", s.handleServeUpload)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.writeError(w, r, notFound("not_found", "no route matches %s %s", r.Method, r.URL.Path))
	})

	return s.withRecovery(s.withRequestID(s.withAccessLog(mux)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	body := map[string]string{"status": "ok"}

	if err := s.store.Ping(ctx); err != nil {
		s.log.ErrorContext(ctx, "health check could not reach the database", slog.String("error", err.Error()))
		body["status"] = "degraded"
		body["error"] = "database unavailable"
		s.writeJSON(w, http.StatusServiceUnavailable, body)
		return
	}
	s.writeJSON(w, http.StatusOK, body)
}
