package hub

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

const (
	maxMessageSize = 512
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
)

type Hub struct {
	Rooms      map[string]*Room
	Clients    map[*Client]bool
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
	MaxPlayers int
}

func NewHub(maxPlayers int) *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		MaxPlayers: maxPlayers,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := h.findAvailableRoom()
	if room == nil {
		roomID := uuid.New().String()
		room = &Room{
			ID:       roomID,
			Name:     fmt.Sprintf("Room %s", roomID[:8]),
			Players:  make([]string, 0),
			Capacity: h.MaxPlayers,
		}
		h.Rooms[roomID] = room
	}

	room.Players = append(room.Players, client.Username)
	client.RoomID = room.ID
	h.Clients[client] = true

	joinMessage, _ := json.Marshal(&Message{
		Type:      "player_joined",
		Room:      room.ID,
		Username:  client.Username,
		Content:   room,
		Timestamp: time.Now().Unix(),
	})

	h.broadcastToRoom(room.ID, joinMessage)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.Clients[client]; ok {
		room := h.Rooms[client.RoomID]
		if room != nil {
			for i, username := range room.Players {
				if username == client.Username {
					room.Players = append(room.Players[:i], room.Players[i+1:]...)
					break
				}
			}

			if len(room.Players) == 0 {
				delete(h.Rooms, client.RoomID)
			} else {
				leaveMessage, _ := json.Marshal(&Message{
					Type:      "player_left",
					Room:      room.ID,
					Username:  client.Username,
					Content:   room,
					Timestamp: time.Now().Unix(),
				})
				h.broadcastToRoom(room.ID, leaveMessage)
			}
		}

		delete(h.Clients, client)
		close(client.Send)
	}
}

func (h *Hub) findAvailableRoom() *Room {
	for _, room := range h.Rooms {
		if len(room.Players) < room.Capacity {
			return room
		}
	}
	return nil
}

func (h *Hub) broadcastMessage(message *Message) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return
	}
	h.broadcastToRoom(message.Room, jsonMessage)
}

func (h *Hub) broadcastToRoom(roomID string, message []byte) {
	for client := range h.Clients {
		if client.RoomID == roomID {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.Clients, client)
			}
		}
	}
}
