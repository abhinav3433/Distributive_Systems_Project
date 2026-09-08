package main

import (
	"database/sql"
	"log"
	"time"

	"order-service/config"
	"order-service/handler"
	"order-service/messaging"
	"order-service/repository"
	"order-service/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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
			shipping_address TEXT,
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

	// 3. Initialize Service & Handler
	svc := service.NewOrderService(repo, publisher, cfg.ProductServiceURL)
	h := handler.NewOrderHandler(svc)

	router := gin.Default()

	// Endpoints
	router.GET("/health", h.HealthCheck)
	router.GET("/orders/health", h.HealthCheck)
	router.POST("/orders", h.CreateOrder)
	router.GET("/orders", h.GetOrders)
	router.GET("/orders/:id", h.GetOrderByID)

	log.Printf("Order Service starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Order Service failed to start: %v", err)
	}
}
