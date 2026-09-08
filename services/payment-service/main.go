package main

import (
	"database/sql"
	"log"
	"time"

	"payment-service/config"
	"payment-service/handler"
	"payment-service/messaging"
	"payment-service/repository"
	"payment-service/service"

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
				log.Println("Successfully connected to PostgreSQL payment_db")
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

func ensureTableExists(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			order_id INT UNIQUE NOT NULL,
			user_id INT NOT NULL,
			amount NUMERIC(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL,
			payment_method VARCHAR(50) DEFAULT 'SIMULATED',
			transaction_id VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	return err
}

func main() {
	cfg := config.LoadConfig()

	// 1. Initialize Database
	db, err := initDB(cfg.GetDSN())
	var repo repository.PaymentRepository

	if err != nil || db == nil {
		log.Printf("Warning: Database connection failed (%v). Falling back to mock payment repository for standalone testing.", err)
		repo = repository.NewMockPaymentRepository()
	} else {
		if err := ensureTableExists(db); err != nil {
			log.Printf("Warning: Table creation failed: %v", err)
		}
		repo = repository.NewPostgresPaymentRepository(db)
	}

	// 2. Initialize Payment Service
	svc := service.NewPaymentService(repo)

	// 3. Initialize & Start RabbitMQ Consumer (auto-reconnecting in background)
	consumer := messaging.NewEventConsumer(cfg.RabbitMQURL, svc)
	consumer.Start()
	defer consumer.Stop()

	// 4. Initialize REST API Handlers
	h := handler.NewPaymentHandler(svc)

	router := gin.Default()

	// REST Endpoints (Preserved as requested)
	router.GET("/health", h.HealthCheck)
	router.GET("/payments/health", h.HealthCheck)
	router.POST("/payments", h.ProcessPayment)
	router.GET("/payments/:orderId", h.GetPaymentByOrderID)

	log.Printf("Payment Service starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Payment Service failed to start: %v", err)
	}
}
