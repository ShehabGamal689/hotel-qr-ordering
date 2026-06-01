package model

import (
	"time"
)

// OrderStatus defines the state of an order
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusAccepted  OrderStatus = "accepted"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

// Property represents a hotel property
type Property struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Room represents a hotel room associated with a property
type Room struct {
	ID         string    `json:"id"`
	PropertyID string    `json:"property_id"`
	RoomNumber string    `json:"room_number"`
	QRToken    string    `json:"qr_token"`
	CreatedAt  time.Time `json:"created_at"`
}

// MenuItem represents a food or beverage offering
type MenuItem struct {
	ID          string    `json:"id"`
	PropertyID  string    `json:"property_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	IsAvailable bool      `json:"is_available"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// Order represents an order placed by a guest in a room
type Order struct {
	ID          string      `json:"id"`
	RoomID      string      `json:"room_id"`
	RoomNumber  string      `json:"room_number"` // Loaded dynamically for convenience
	Status      OrderStatus `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	CreatedAt   time.Time   `json:"created_at"`
	Items       []OrderItem `json:"items,omitempty"`
}

// OrderItem represents a single item and quantity inside an order
type OrderItem struct {
	ID         string   `json:"id"`
	OrderID    string   `json:"order_id"`
	MenuItemID string   `json:"menu_item_id"`
	ItemName   string   `json:"item_name,omitempty"` // Loaded dynamically for convenience
	Quantity   int      `json:"quantity"`
	Price      float64  `json:"price"` // Price at time of ordering
}

// OrderItemRequest is used to parse incoming items in OrderRequest
type OrderItemRequest struct {
	MenuItemID string `json:"menu_item_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
}

// OrderRequest represents the body sent by clients to place an order
type OrderRequest struct {
	RoomNumber string             `json:"room_number" binding:"required"`
	Items      []OrderItemRequest `json:"items" binding:"required,dive"`
}

// StatusUpdateRequest represents the body sent by admin to update order status
type StatusUpdateRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}

// WSEvent represents the payload broadcast over WebSockets
type WSEvent struct {
	Type    string      `json:"type"` // e.g. "order_created", "order_updated"
	Payload interface{} `json:"payload"`
}
