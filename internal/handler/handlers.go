package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adham/hotel-qr-ordering/internal/model"
	"github.com/adham/hotel-qr-ordering/internal/service"
)

type HTTPHandler struct {
	srv *service.HotelService
	hub *WSHub
}

func NewHTTPHandler(srv *service.HotelService, hub *WSHub) *HTTPHandler {
	return &HTTPHandler{
		srv: srv,
		hub: hub,
	}
}

// GetMenu handles GET /api/v1/menu?room=NUMBER
func (h *HTTPHandler) GetMenu(c *gin.Context) {
	roomNumber := c.Query("room")
	if roomNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room query parameter is required"})
		return
	}

	items, room, err := h.srv.GetMenuForRoom(c.Request.Context(), roomNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room":       room,
		"menu_items": items,
	})
}

// CreateOrder handles POST /api/v1/orders
func (h *HTTPHandler) CreateOrder(c *gin.Context) {
	var req model.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.srv.PlaceOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrders handles GET /api/v1/orders
func (h *HTTPHandler) GetOrders(c *gin.Context) {
	orders, err := h.srv.GetActiveOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus handles PATCH /api/v1/orders/:id/status
func (h *HTTPHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID path parameter is required"})
		return
	}

	var req model.StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.srv.UpdateOrderStatus(c.Request.Context(), orderID, req.Status)
	if err != nil {
		if errors.Is(err, errors.New("not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ServeOrderPage serves the guest mobile order HTML page
func (h *HTTPHandler) ServeOrderPage(c *gin.Context) {
	c.File("./web/static/order.html")
}

// ServeAdminPage serves the admin dashboard HTML page
func (h *HTTPHandler) ServeAdminPage(c *gin.Context) {
	c.File("./web/static/admin.html")
}
