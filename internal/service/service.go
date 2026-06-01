package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/adham/hotel-qr-ordering/internal/model"
	"github.com/adham/hotel-qr-ordering/internal/repository"
)

type HotelService struct {
	dbRepo    *repository.PostgresRepository
	redisRepo *repository.RedisRepository
}

func NewHotelService(db *repository.PostgresRepository, redis *repository.RedisRepository) *HotelService {
	return &HotelService{
		dbRepo:    db,
		redisRepo: redis,
	}
}

// GetRoomByNumber retrieves room info
func (s *HotelService) GetRoomByNumber(ctx context.Context, roomNumber string) (*model.Room, error) {
	return s.dbRepo.GetRoomByNumber(ctx, roomNumber)
}

// GetRoomByToken retrieves room info via QR token
func (s *HotelService) GetRoomByToken(ctx context.Context, token string) (*model.Room, error) {
	return s.dbRepo.GetRoomByToken(ctx, token)
}

// GetMenuForRoom retrieves room and its property's menu (with Redis caching)
func (s *HotelService) GetMenuForRoom(ctx context.Context, roomNumber string) ([]model.MenuItem, *model.Room, error) {
	room, err := s.dbRepo.GetRoomByNumber(ctx, roomNumber)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch room: %w", err)
	}

	// Try reading from Redis Cache first
	items, err := s.redisRepo.GetCachedMenu(ctx, room.PropertyID)
	if err != nil {
		log.Printf("Warning: Failed to get cached menu from Redis: %v", err)
	}

	if items != nil {
		return items, room, nil
	}

	// Cache miss: query PostgreSQL
	items, err = s.dbRepo.GetMenuItems(ctx, room.PropertyID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch menu items from db: %w", err)
	}

	// Cache items in Redis for 1 hour
	if err := s.redisRepo.CacheMenu(ctx, room.PropertyID, items, 1*time.Hour); err != nil {
		log.Printf("Warning: Failed to cache menu items in Redis: %v", err)
	}

	return items, room, nil
}

// PlaceOrder validates and creates a new room order
func (s *HotelService) PlaceOrder(ctx context.Context, req *model.OrderRequest) (*model.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("cannot place an order with empty items")
	}

	// Find the room
	room, err := s.dbRepo.GetRoomByNumber(ctx, req.RoomNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid room number: %w", err)
	}

	var totalAmount float64
	var orderItems []model.OrderItem

	// Validate items and calculate total amount
	for _, reqItem := range req.Items {
		menuItem, err := s.dbRepo.GetMenuItemByID(ctx, reqItem.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("menu item %s not found: %w", reqItem.MenuItemID, err)
		}

		if !menuItem.IsAvailable {
			return nil, fmt.Errorf("menu item '%s' is currently unavailable", menuItem.Name)
		}

		if menuItem.PropertyID != room.PropertyID {
			return nil, fmt.Errorf("menu item '%s' does not belong to room property", menuItem.Name)
		}

		itemPrice := menuItem.Price
		itemTotal := itemPrice * float64(reqItem.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, model.OrderItem{
			MenuItemID: menuItem.ID,
			ItemName:   menuItem.Name,
			Quantity:   reqItem.Quantity,
			Price:      itemPrice,
		})
	}

	// Construct order entity
	order := &model.Order{
		RoomID:      room.ID,
		RoomNumber:  room.RoomNumber,
		TotalAmount: totalAmount,
		Items:       orderItems,
	}

	// Write to database in transaction
	if err := s.dbRepo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Publish real-time order created event to Redis Pub/Sub
	if err := s.redisRepo.PublishOrderEvent(ctx, "order_created", order); err != nil {
		log.Printf("Warning: Failed to publish order_created event to Redis: %v", err)
	}

	return order, nil
}

// GetActiveOrders lists all active hotel orders for receptionist
func (s *HotelService) GetActiveOrders(ctx context.Context) ([]model.Order, error) {
	return s.dbRepo.GetOrders(ctx)
}

// UpdateOrderStatus updates order status and publishes websocket update event
func (s *HotelService) UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) (*model.Order, error) {
	// Validate status transition values
	switch status {
	case model.StatusPending, model.StatusAccepted, model.StatusCompleted, model.StatusCancelled:
		// valid
	default:
		return nil, fmt.Errorf("invalid order status value: %s", status)
	}

	// Perform database update
	if err := s.dbRepo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		return nil, fmt.Errorf("failed to update status in database: %w", err)
	}

	// Fetch updated order to get room numbers and order items
	order, err := s.dbRepo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated order details: %w", err)
	}

	// Publish real-time update notification to Redis Pub/Sub
	if err := s.redisRepo.PublishOrderEvent(ctx, "order_updated", order); err != nil {
		log.Printf("Warning: Failed to publish order_updated event to Redis: %v", err)
	}

	return order, nil
}
