package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/adham/hotel-qr-ordering/internal/auth"
	"github.com/adham/hotel-qr-ordering/internal/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Enforce safety: Origin validation can be whitelisted. For dev, we allow any origin.
		return true
	},
}

// Client represents a connected admin dashboard or guest portal websocket connection
type Client struct {
	hub        *WSHub
	conn       *websocket.Conn
	send       chan []byte
	propertyID string // Used to route messages only to this client's tenant
	roomID     string // Filter for guest specific updates if applicable
	clientType string // "admin" or "client"
}

// WSHub maintains active clients and broadcasts events subscribed from Redis
type WSHub struct {
	redisRepo  *repository.RedisRepository
	dbRepo     *repository.PostgresRepository
	clients    map[string]map[*Client]bool // Mapped by propertyID to isolate tenants
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewWSHub(redis *repository.RedisRepository, db *repository.PostgresRepository) *WSHub {
	return &WSHub{
		redisRepo:  redis,
		dbRepo:     db,
		clients:    make(map[string]map[*Client]bool),
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
				if h.clients[client.propertyID] == nil {
					h.clients[client.propertyID] = make(map[*Client]bool)
				}
				h.clients[client.propertyID][client] = true
				h.mu.Unlock()
				log.Printf("WebSocket: New %s registered for property %s", client.clientType, client.propertyID)

			case client := <-h.unregister:
				h.mu.Lock()
				if cls, ok := h.clients[client.propertyID]; ok {
					if _, okClient := cls[client]; okClient {
						delete(cls, client)
						close(client.send)
						log.Printf("WebSocket: %s unregistered for property %s", client.clientType, client.propertyID)
					}
					if len(cls) == 0 {
						delete(h.clients, client.propertyID)
					}
				}
				h.mu.Unlock()
			}
		}
	}()

	// 2. Start Redis Pub/Sub listener and broadcast incoming messages to clients of target property
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
				h.broadcastByProperty([]byte(msg.Payload))
			}
		}
	}()
}

func (h *WSHub) broadcastByProperty(message []byte) {
	var rawEvent struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(message, &rawEvent); err != nil {
		log.Printf("WebSocket Hub: unmarshal event failure: %v", err)
		return
	}

	var innerPayload struct {
		PropertyID string `json:"property_id"`
	}
	_ = json.Unmarshal(rawEvent.Payload, &innerPayload)

	propertyID := innerPayload.PropertyID

	if propertyID == "" {
		// Fallback for orders which embed property_id inside
		var orderPayload struct {
			PropertyID string `json:"property_id"`
		}
		_ = json.Unmarshal(rawEvent.Payload, &orderPayload)
		propertyID = orderPayload.PropertyID
	}

	if propertyID == "" {
		log.Printf("WebSocket Hub: event type '%s' is missing property_id, skipping broadcast routing.", rawEvent.Type)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if cls, ok := h.clients[propertyID]; ok {
		for client := range cls {
			select {
			case client.send <- message:
			default:
				log.Println("WebSocket Hub: Client channel full, dropping client connection")
				go h.cleanupClient(client)
			}
		}
	}
}

func (h *WSHub) cleanupClient(client *Client) {
	client.conn.Close()
	h.unregister <- client
}

func (h *WSHub) ServeWS(c *gin.Context) {
	var propertyID string
	var roomID string
	var clientType string

	path := c.Request.URL.Path
	switch path {
	case "/ws/admin":
		clientType = "admin"
		token := c.Query("token")
		if token == "" {
			if cookieToken, err := c.Cookie("admin_token"); err == nil {
				token = cookieToken
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: admin token required"})
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
			return
		}
		propertyID = claims.PropertyID
	case "/ws/client":
		clientType = "client"
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: room token required"})
			return
		}

		room, err := h.dbRepo.GetRoomByToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("invalid room token: %v", err)})
			return
		}
		propertyID = room.PropertyID
		roomID = room.ID
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid websocket endpoint"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket connection: %v", err)
		return
	}

	client := &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, 256),
		propertyID: propertyID,
		roomID:     roomID,
		clientType: clientType,
	}

	h.register <- client

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
