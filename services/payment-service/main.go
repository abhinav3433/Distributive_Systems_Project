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

	"payment-service/config"
	paymentgrpc "payment-service/grpc"
	"payment-service/handler"
	"payment-service/messaging"
	"payment-service/repository"
	"payment-service/service"
	paymentv1 "proto/payment/v1"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	googlegrpc "google.golang.org/grpc"
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

	// 5. Initialize and start Payment Service gRPC Server
	grpcServer := googlegrpc.NewServer()
	paymentGRPCServer := paymentgrpc.NewPaymentGRPCServer(svc)
	paymentv1.RegisterPaymentServiceServer(grpcServer, paymentGRPCServer)

	// Register standard gRPC health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("payment.v1.PaymentService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("Payment Service gRPC server listening on port %s...", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Payment Service gRPC server error: %v", err)
		}
	}()

	// 6. HTTP Server (REST + Health)
	router := gin.Default()

	// Root endpoint (prevents 404 when tested directly)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "payment-service",
			"status":    "UP",
			"http_port": cfg.Port,
			"grpc_port": cfg.GRPCPort,
			"endpoints": []string{"/health", "/payments", "/payments/:orderId"},
		})
	})

	// REST Endpoints (Preserved as requested)
	router.GET("/health", h.HealthCheck)
	router.GET("/payments/health", h.HealthCheck)
	router.GET("/api/payments/health", h.HealthCheck)
	router.POST("/payments", h.ProcessPayment)
	router.POST("/api/payments", h.ProcessPayment)
	router.GET("/payments/:orderId", h.GetPaymentByOrderID)
	router.GET("/api/payments/:orderId", h.GetPaymentByOrderID)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Payment Service HTTP server listening on port %s...", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Payment Service HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Payment Service...")

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}
	log.Println("Payment Service gracefully stopped.")
}
