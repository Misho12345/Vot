package hub

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
	RoomID   string
	mu       sync.Mutex
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		err := c.Conn.Close()
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return err
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.Hub.Broadcast <- &Message{
			Type:      "message",
			Room:      c.RoomID,
			Username:  c.Username,
			Content:   string(message),
			Timestamp: time.Now().Unix(),
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			if !ok {
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("error: %v", err)
					return
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, err = w.Write(message)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
