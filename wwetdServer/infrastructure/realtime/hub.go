package realtime

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	upgrader   websocket.Upgrader
}

func NewHub(allowedOrigins []string) *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     buildOriginChecker(allowedOrigins),
		},
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				close(client.send)
				_ = client.conn.Close()
				delete(h.clients, client)
			}
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := newClient(h, conn)
	h.register <- client

	go client.writePump()
	go client.readPump()

	return nil
}

func buildOriginChecker(allowedOrigins []string) func(r *http.Request) bool {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			allowAll = true
			break
		}
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(r *http.Request) bool {
		if allowAll {
			return true
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		_, ok := allowed[origin]
		return ok
	}
}
