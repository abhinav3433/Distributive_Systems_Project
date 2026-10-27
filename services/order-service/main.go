package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/config"
	ordergrpc "order-service/grpc"
	"order-service/handler"
	"order-service/messaging"
	"order-service/repository"
	"order-service/service"
	orderv1 "proto/order/v1"
	productv1 "proto/product/v1"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func initDB(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var lastErr error

	for i := 0; i < 3; i++ {
		var err error
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			if pingErr := db.Ping(); pingErr == nil {
				log.Println("Successfully connected to PostgreSQL order_db")
				return db, nil
			} else {
				lastErr = pingErr
			}
		} else {
			lastErr = err
		}
		log.Printf("Waiting for PostgreSQL connection (attempt %d/3)...", i+1)
		time.Sleep(1 * time.Second)
	}

	if db != nil {
		db.Close()
	}
	return nil, lastErr
}

func ensureTablesExist(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
			total_amount NUMERIC(10, 2) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INT REFERENCES orders(id) ON DELETE CASCADE,
			product_id INT NOT NULL,
			product_name VARCHAR(255),
			quantity INT NOT NULL,
			price NUMERIC(10, 2) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS order_shipping_addresses (
			id SERIAL PRIMARY KEY,
			order_id INT UNIQUE NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			recipient_name VARCHAR(255) NOT NULL,
			phone_number VARCHAR(50) NOT NULL,
			address_line1 VARCHAR(255) NOT NULL,
			address_line2 VARCHAR(255),
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100) NOT NULL,
			postal_code VARCHAR(50) NOT NULL,
			country VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
		CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
		CREATE INDEX IF NOT EXISTS idx_order_shipping_addresses_order_id ON order_shipping_addresses(order_id);
	`
	_, err := db.Exec(query)
	return err
}

func main() {
	cfg := config.LoadConfig()

	// 1. Initialize Database
	db, err := initDB(cfg.GetDSN())
	var repo repository.OrderRepository

	if err != nil || db == nil {
		log.Printf("Warning: Database connection failed (%v). Falling back to mock order repository for standalone testing.", err)
		repo = repository.NewMockOrderRepository()
	} else {
		if err := ensureTablesExist(db); err != nil {
			log.Printf("Warning: Table creation failed: %v", err)
		}
		repo = repository.NewPostgresOrderRepository(db)
	}

	// 2. Initialize RabbitMQ Publisher
	var publisher messaging.EventPublisher
	var rmqPub *messaging.RabbitMQPublisher
	for i := 0; i < 10; i++ {
		rmqPub, err = messaging.NewRabbitMQPublisher(cfg.RabbitMQURL)
		if err == nil {
			break
		}
		log.Printf("Waiting for RabbitMQ connection (attempt %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Printf("Warning: RabbitMQ connection failed (%v). Falling back to mock event publisher.", err)
		publisher = messaging.NewMockEventPublisher()
	} else {
		log.Printf("Successfully connected to RabbitMQ at %s", cfg.RabbitMQURL)
		publisher = rmqPub
		defer rmqPub.Close()
	}

	// 3. Initialize gRPC connection to Product Service (if reachable)
	var productGRPCClient productv1.ProductServiceClient
	productConn, err := googlegrpc.Dial(cfg.ProductServiceGRPCAddr,
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Warning: Could not connect to Product Service via gRPC at %s: %v. Will fall back to HTTP.", cfg.ProductServiceGRPCAddr, err)
	} else {
		log.Printf("Successfully established gRPC client to Product Service at %s", cfg.ProductServiceGRPCAddr)
		productGRPCClient = productv1.NewProductServiceClient(productConn)
		defer productConn.Close()
	}

	// 4. Initialize Service & Handler
	svc := service.NewOrderServiceWithGRPC(repo, publisher, cfg.ProductServiceURL, productGRPCClient)
	h := handler.NewOrderHandler(svc)

	// 5. Initialize and start Order Service gRPC Server
	grpcServer := googlegrpc.NewServer()
	orderGRPCServer := ordergrpc.NewOrderGRPCServer(svc)
	orderv1.RegisterOrderServiceServer(grpcServer, orderGRPCServer)

	// Register standard gRPC health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("order.v1.OrderService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("Order Service gRPC server listening on port %s...", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Order Service gRPC server error: %v", err)
		}
	}()

	// 6. HTTP Server (REST + Health)
	router := gin.Default()

	// Root endpoint (prevents 404 when tested directly)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "order-service",
			"status":    "UP",
			"http_port": cfg.Port,
			"grpc_port": cfg.GRPCPort,
			"endpoints": []string{"/health", "/orders", "/orders/:id"},
		})
	})

	// Routes
	router.GET("/health", h.HealthCheck)
	router.GET("/orders/health", h.HealthCheck)
	router.GET("/api/orders/health", h.HealthCheck)
	router.POST("/orders", h.CreateOrder)
	router.POST("/api/orders", h.CreateOrder)
	router.GET("/orders", h.GetOrders)
	router.GET("/api/orders", h.GetOrders)
	router.GET("/orders/:id", h.GetOrderByID)
	router.GET("/api/orders/:id", h.GetOrderByID)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Order Service HTTP server listening on port %s...", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Order Service HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Order Service...")

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}
	log.Println("Order Service gracefully stopped.")
}
