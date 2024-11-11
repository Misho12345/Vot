package handlers

import (
	"backend/internal/auth"
	"backend/internal/hub"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				return true
			}
		}
		return false
	},
}

type RoomHandler struct {
	hub *hub.Hub
}

func NewRoomHandler(maxPlayers int) *RoomHandler {
	return &RoomHandler{
		hub: hub.NewHub(maxPlayers),
	}
}

func (h *RoomHandler) Start() {
	go h.hub.Run()
}

func (h *RoomHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		http.Error(w, "Token required", http.StatusBadRequest)
		return
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}

	client := &hub.Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Username: claims.Username,
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
