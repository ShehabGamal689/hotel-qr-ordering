package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/adham/hotel-qr-ordering/internal/handler"
	"github.com/adham/hotel-qr-ordering/internal/repository"
	"github.com/adham/hotel-qr-ordering/internal/service"
)

func main() {
	log.Println("Starting Hotel QR Ordering Backend...")

	// 1. Load Configurations from Env with defaults
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/hotel_ordering?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	migrationsDir := getEnv("MIGRATIONS_DIR", "./db")

	// 2. Set up Root Context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Initialize Postgres Repository & Run Migrations/Seeds
	dbRepo, err := repository.NewPostgresRepository(dbURL)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}
	defer dbRepo.Close()

	if err := dbRepo.RunMigrations(migrationsDir); err != nil {
		log.Fatalf("Fatal: Database migrations failed: %v", err)
	}

	// 4. Initialize Redis Repository
	redisRepo, err := repository.NewRedisRepository(redisURL, "", 0)
	if err != nil {
		log.Fatalf("Fatal: Redis initialization failed: %v", err)
	}
	defer func() {
		if err := redisRepo.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
	}()

	// 5. Initialize & Start WebSocket Hub
	hub := handler.NewWSHub(redisRepo)
	hub.Start(ctx)

	// 6. Initialize Service Layer & Handler Layer
	srv := service.NewHotelService(dbRepo, redisRepo)
	h := handler.NewHTTPHandler(srv, hub)

	// 7. Setup Gin Server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// CORS Middleware (Simple and clean, supporting all methods and content headers)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Serve SPAs statically
	r.GET("/order", h.ServeOrderPage)
	r.GET("/admin", h.ServeAdminPage)
	r.Static("/static", "./web/static")

	// Redirect root to Room 101 Order Page
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/order?room=101")
	})

	// API Routes Group
	api := r.Group("/api/v1")
	{
		api.GET("/menu", h.GetMenu)
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders", h.GetOrders)
		api.PATCH("/orders/:id/status", h.UpdateOrderStatus)
	}

	// WebSocket upgrading endpoint
	r.GET("/ws/admin", hub.ServeWS)

	// 8. Start HTTP Server with Graceful Shutdown
	srvAddr := ":" + port
	server := &http.Server{
		Addr:    srvAddr,
		Handler: r,
	}

	go func() {
		log.Printf("Server listening on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Fatal: Listen and serve failed: %v", err)
		}
	}()

	// Wait for OS shutdown signal
	<-ctx.Done()
	log.Println("Shutdown signal received, shutting down server gracefully...")

	// Context for server graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Fatal: Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped successfully")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
