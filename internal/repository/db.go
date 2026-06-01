package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adham/hotel-qr-ordering/internal/model"
)

type PostgresRepository struct {
	Pool *pgxpool.Pool
}

// NewPostgresRepository establishes a connection pool to Postgres
func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Adjust pool settings
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Ping to ensure connection works
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return &PostgresRepository{Pool: pool}, nil
}

// Close closes the connection pool
func (r *PostgresRepository) Close() {
	if r.Pool != nil {
		r.Pool.Close()
	}
}

// RunMigrations reads and runs schema.sql and seed.sql from the given directory
func (r *PostgresRepository) RunMigrations(migrationsDir string) error {
	ctx := context.Background()

	// Read and execute schema
	schemaPath := filepath.Join(migrationsDir, "schema.sql")
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	log.Printf("Executing schema from %s...", schemaPath)
	if _, err := r.Pool.Exec(ctx, string(schemaSQL)); err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}
	log.Println("Schema migration completed successfully")

	// Read and execute seed data
	seedPath := filepath.Join(migrationsDir, "seed.sql")
	seedSQL, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed.sql: %w", err)
	}

	log.Printf("Executing seed data from %s...", seedPath)
	if _, err := r.Pool.Exec(ctx, string(seedSQL)); err != nil {
		return fmt.Errorf("failed to execute seed.sql: %w", err)
	}
	log.Println("Database seeding completed successfully")

	return nil
}

// GetRoomByNumber retrieves a room by its room number
func (r *PostgresRepository) GetRoomByNumber(ctx context.Context, roomNumber string) (*model.Room, error) {
	query := `SELECT id, property_id, room_number, qr_token, created_at FROM rooms WHERE room_number = $1`
	var rm model.Room
	err := r.Pool.QueryRow(ctx, query, roomNumber).Scan(&rm.ID, &rm.PropertyID, &rm.RoomNumber, &rm.QRToken, &rm.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("room %s not found", roomNumber)
		}
		return nil, err
	}
	return &rm, nil
}

// GetRoomByToken retrieves a room by its QR token
func (r *PostgresRepository) GetRoomByToken(ctx context.Context, token string) (*model.Room, error) {
	query := `SELECT id, property_id, room_number, qr_token, created_at FROM rooms WHERE qr_token = $1`
	var rm model.Room
	err := r.Pool.QueryRow(ctx, query, token).Scan(&rm.ID, &rm.PropertyID, &rm.RoomNumber, &rm.QRToken, &rm.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("room with token %s not found", token)
		}
		return nil, err
	}
	return &rm, nil
}

// GetMenuItems retrieves all menu items for a given property
func (r *PostgresRepository) GetMenuItems(ctx context.Context, propertyID string) ([]model.MenuItem, error) {
	query := `SELECT id, property_id, name, description, price, is_available, category, created_at 
	          FROM menu_items WHERE property_id = $1 ORDER BY category, name`
	rows, err := r.Pool.Query(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.MenuItem
	for rows.Next() {
		var item model.MenuItem
		err := rows.Scan(
			&item.ID,
			&item.PropertyID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.IsAvailable,
			&item.Category,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetMenuItemByID retrieves a single menu item by ID
func (r *PostgresRepository) GetMenuItemByID(ctx context.Context, itemID string) (*model.MenuItem, error) {
	query := `SELECT id, property_id, name, description, price, is_available, category, created_at 
	          FROM menu_items WHERE id = $1`
	var item model.MenuItem
	err := r.Pool.QueryRow(ctx, query, itemID).Scan(
		&item.ID,
		&item.PropertyID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.IsAvailable,
		&item.Category,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// CreateOrder inserts an order and its items within a database transaction
func (r *PostgresRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert order
	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.Status = model.StatusPending

	orderQuery := `INSERT INTO orders (id, room_id, status, total_amount, created_at) 
	               VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, orderQuery, order.ID, order.RoomID, order.Status, order.TotalAmount, order.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	itemQuery := `INSERT INTO order_items (id, order_id, menu_item_id, quantity, price) 
	              VALUES ($1, $2, $3, $4, $5)`
	for i := range order.Items {
		item := &order.Items[i]
		item.ID = uuid.New().String()
		item.OrderID = order.ID

		_, err = tx.Exec(ctx, itemQuery, item.ID, item.OrderID, item.MenuItemID, item.Quantity, item.Price)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetOrders retrieves all orders with details
func (r *PostgresRepository) GetOrders(ctx context.Context) ([]model.Order, error) {
	// First fetch orders and join with rooms to get room number
	orderQuery := `
		SELECT o.id, o.room_id, r.room_number, o.status, o.total_amount, o.created_at 
		FROM orders o
		JOIN rooms r ON o.room_id = r.id
		ORDER BY o.created_at DESC`
	
	rows, err := r.Pool.Query(ctx, orderQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		err := rows.Scan(&o.ID, &o.RoomID, &o.RoomNumber, &o.Status, &o.TotalAmount, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	// Fetch items for each order
	for i := range orders {
		items, err := r.getOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}

	return orders, nil
}

// GetOrder retrieves a single order by ID
func (r *PostgresRepository) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	query := `
		SELECT o.id, o.room_id, r.room_number, o.status, o.total_amount, o.created_at 
		FROM orders o
		JOIN rooms r ON o.room_id = r.id
		WHERE o.id = $1`
	
	var o model.Order
	err := r.Pool.QueryRow(ctx, query, orderID).Scan(&o.ID, &o.RoomID, &o.RoomNumber, &o.Status, &o.TotalAmount, &o.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order %s not found", orderID)
		}
		return nil, err
	}

	items, err := r.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

// UpdateOrderStatus transitions an order's status
func (r *PostgresRepository) UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	tag, err := r.Pool.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("order %s not found", orderID)
	}
	return nil
}

// Helper to fetch order items with menu item names
func (r *PostgresRepository) getOrderItems(ctx context.Context, orderID string) ([]model.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.menu_item_id, mi.name, oi.quantity, oi.price 
		FROM order_items oi
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		WHERE oi.order_id = $1`
	
	rows, err := r.Pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.ItemName, &item.Quantity, &item.Price)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
