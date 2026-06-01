package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/adham/hotel-qr-ordering/internal/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connection from any origin for local testing
		return true
	},
}

// Client represents a connected admin dashboard websocket connection
type Client struct {
	hub  *WSHub
	conn *websocket.Conn
	send chan []byte
}

// WSHub maintains active clients and broadcasts events subscribed from Redis
type WSHub struct {
	redisRepo  *repository.RedisRepository
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewWSHub(redis *repository.RedisRepository) *WSHub {
	return &WSHub{
		redisRepo:  redis,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Start runs the hub event loop and starts the Redis Pub/Sub subscription listener
func (h *WSHub) Start(ctx context.Context) {
	// 1. Run the register/unregister loop
	go func() {
		for {
			select {
			case client := <-h.register:
				h.mu.Lock()
				h.clients[client] = true
				h.mu.Unlock()
				log.Println("New WebSocket client registered")

			case client := <-h.unregister:
				h.mu.Lock()
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					log.Println("WebSocket client unregistered")
				}
				h.mu.Unlock()
			}
		}
	}()

	// 2. Start Redis Pub/Sub listener and broadcast incoming messages to clients
	go func() {
		pubsub := h.redisRepo.SubscribeToOrderEvents(ctx)
		defer pubsub.Close()

		ch := pubsub.Channel()
		log.Println("WebSocket Hub listening to Redis Pub/Sub channel")

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping Redis Pub/Sub listener due to context cancellation")
				return
			case msg, ok := <-ch:
				if !ok {
					log.Println("Redis Pub/Sub channel closed")
					return
				}

				log.Printf("Received message from Redis: %s", msg.Payload)
				h.broadcast([]byte(msg.Payload))
			}
		}
	}()
}

// broadcast sends a raw message to all registered clients
func (h *WSHub) broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			log.Println("Client channel full, dropping client")
			go h.cleanupClient(client)
		}
	}
}

func (h *WSHub) cleanupClient(client *Client) {
	client.conn.Close()
	h.unregister <- client
}

// ServeWS upgrades the HTTP request to WebSocket and registers the client
func (h *WSHub) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket connection: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not upgrade connection"})
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	// Start reading and writing pumps
	go client.writePump()
	go client.readPump()
}

// writePump pushes messages from client.send channel to WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump listens for incoming messages from websocket client (mostly for checking connectivity/pongs)
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Process client heartbeat or incoming actions (none required for simple dashboard reader, but could handle ping-pong)
		var ping struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(message, &ping); err == nil && ping.Type == "ping" {
			c.send <- []byte(`{"type":"pong"}`)
		}
	}
}
