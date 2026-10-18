package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"bowatt/internal/hub"
	"bowatt/internal/store"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 30 * time.Second
)

type wsEnvelope struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// handleWebSocket upgrades the connection and streams new images.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		s.log.WarnContext(r.Context(), "websocket upgrade refused",
			slog.String("error", err.Error()), slog.String("remote", r.RemoteAddr))
		return
	}
	defer conn.CloseNow()

	client := hub.NewClient(hub.ClientBuffer)
	s.hub.Register(client)
	defer s.hub.Unregister(client)

	// CloseRead handles incoming control frames and cancels ctx when the peer
	// disconnects.
	ctx := conn.CloseRead(r.Context())

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-client.Messages():
			if !ok {
				_ = conn.Close(websocket.StatusTryAgainLater, "reconnect to resynchronise")
				return
			}
			if err := writeFrame(ctx, conn, message); err != nil {
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func writeFrame(ctx context.Context, conn *websocket.Conn, message []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, message)
}

// broadcastImage pushes a newly stored image to every connected client.
func (s *Server) broadcastImage(image store.Image) {
	payload, err := json.Marshal(wsEnvelope{Type: "image.created", Data: image})
	if err != nil {
		s.log.Error("could not encode broadcast payload", slog.String("error", err.Error()))
		return
	}
	s.hub.Broadcast(payload)
}
