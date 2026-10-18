package hub

import (
	"log/slog"
	"sync"
)

const ClientBuffer = 16

type Client struct {
	send      chan []byte
	closeOnce sync.Once
}

func NewClient(buffer int) *Client {
	return &Client{send: make(chan []byte, buffer)}
}

// Messages is the queue the connection writer drains. A closed channel means
// this client was dropped.
func (c *Client) Messages() <-chan []byte {
	return c.send
}

func (c *Client) close() {
	c.closeOnce.Do(func() { close(c.send) })
}

func (c *Client) trySend(msg []byte) bool {
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
}

type Hub struct {
	log *slog.Logger

	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	done     chan struct{}
	doneOnce sync.Once
	stopped  chan struct{}

	clients map[*Client]struct{}
}

func New(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	return &Hub{
		log:        log,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 64),
		done:       make(chan struct{}),
		stopped:    make(chan struct{}),
		clients:    make(map[*Client]struct{}),
	}
}

func (h *Hub) Run() {
	defer close(h.stopped)

	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
			h.log.Debug("websocket client connected", "clients", len(h.clients))
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				c.close()
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				if !c.trySend(msg) {
					delete(h.clients, c)
					c.close()
					h.log.Warn("dropped a websocket client that could not keep up")
				}
			}
		case <-h.done:
			for c := range h.clients {
				delete(h.clients, c)
				c.close()
			}
			return
		}
	}
}

func (h *Hub) Register(c *Client) {
	select {
	case h.register <- c:
	case <-h.done:
		c.close()
	}
}

func (h *Hub) Unregister(c *Client) {
	select {
	case h.unregister <- c:
	case <-h.done:
		c.close()
	}
}

func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	case <-h.done:
	}
}

func (h *Hub) Shutdown() {
	h.doneOnce.Do(func() { close(h.done) })
}

func (h *Hub) Wait() {
	<-h.stopped
}
