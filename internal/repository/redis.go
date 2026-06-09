package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/adham/hotel-qr-ordering/internal/model"
)

const (
	MenuCacheKeyPrefix = "menu:property:"
	OrderEventsChannel = "order_events"
)

type RedisRepository struct {
	Client *redis.Client
}

// NewRedisRepository creates a new Redis client wrapper
func NewRedisRepository(addr string, password string, db int) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Println("Successfully connected to Redis")
	return &RedisRepository{Client: client}, nil
}

// Close closes the Redis client
func (r *RedisRepository) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

// CacheMenu serializes and caches catalog items for a property
func (r *RedisRepository) CacheMenu(ctx context.Context, propertyID string, items []model.CatalogItem, ttl time.Duration) error {
	key := MenuCacheKeyPrefix + propertyID
	data, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal catalog items for cache: %w", err)
	}

	err = r.Client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to write catalog items to redis: %w", err)
	}

	log.Printf("Cached %d catalog items for property %s in Redis (TTL: %s)", len(items), propertyID, ttl)
	return nil
}

// GetCachedMenu retrieves and deserializes catalog items for a property from Redis
func (r *RedisRepository) GetCachedMenu(ctx context.Context, propertyID string) ([]model.CatalogItem, error) {
	key := MenuCacheKeyPrefix + propertyID
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss, not an error
		}
		return nil, fmt.Errorf("failed to read from redis: %w", err)
	}

	var items []model.CatalogItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached catalog items: %w", err)
	}

	return items, nil
}

// InvalidateMenuCache invalidates cached catalog items for a property
func (r *RedisRepository) InvalidateMenuCache(ctx context.Context, propertyID string) error {
	key := MenuCacheKeyPrefix + propertyID
	return r.Client.Del(ctx, key).Err()
}

// PublishOrderEvent publishes a real-time event to the Redis Pub/Sub channel
func (r *RedisRepository) PublishOrderEvent(ctx context.Context, eventType string, payload interface{}) error {
	event := model.WSEvent{
		Type:    eventType,
		Payload: payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ws event: %w", err)
	}

	err = r.Client.Publish(ctx, OrderEventsChannel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish event to redis: %w", err)
	}

	log.Printf("Published event '%s' to Redis Pub/Sub channel '%s'", eventType, OrderEventsChannel)
	return nil
}

// SubscribeToOrderEvents returns a Redis Pub/Sub subscription for order events
func (r *RedisRepository) SubscribeToOrderEvents(ctx context.Context) *redis.PubSub {
	pubsub := r.Client.Subscribe(ctx, OrderEventsChannel)
	return pubsub
}
