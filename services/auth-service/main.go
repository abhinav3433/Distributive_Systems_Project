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

	"auth-service/config"
	authgrpc "auth-service/grpc"
	"auth-service/handler"
	"auth-service/middleware"
	"auth-service/repository"
	"auth-service/service"
	authv1 "proto/auth/v1"

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
				log.Println("Successfully connected to PostgreSQL database")
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
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	return err
}

func main() {
	cfg := config.LoadConfig()

	db, err := initDB(cfg.GetDSN())
	var userRepo repository.UserRepository

	if err != nil || db == nil {
		log.Printf("Warning: Database connection failed (%v). Falling back to mock repository for local standalone testing.", err)
		userRepo = repository.NewMockUserRepository()
	} else {
		if err := ensureTableExists(db); err != nil {
			log.Printf("Warning: Auto-migration failed: %v", err)
		}
		userRepo = repository.NewPostgresUserRepository(db)
	}

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	// 1. Initialize and start gRPC Server
	grpcServer := googlegrpc.NewServer()
	authGRPCServer := authgrpc.NewAuthGRPCServer(authSvc)
	authv1.RegisterAuthServiceServer(grpcServer, authGRPCServer)

	// Register standard gRPC health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("auth.v1.AuthService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("Auth Service gRPC server listening on port %s...", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Auth Service gRPC server error: %v", err)
		}
	}()

	// 2. HTTP Server (REST + Health)
	router := gin.Default()

	// Root endpoint (prevents 404 when tested directly)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "auth-service",
			"status":    "UP",
			"http_port": cfg.Port,
			"grpc_port": cfg.GRPCPort,
			"endpoints": []string{"/health", "/register", "/login", "/me"},
		})
	})

	// Routes
	router.GET("/health", authHandler.HealthCheck)
	router.GET("/auth/health", authHandler.HealthCheck)
	router.GET("/api/auth/health", authHandler.HealthCheck)
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(authSvc))
	{
		protected.GET("/me", authHandler.GetMe)
		protected.GET("/auth/me", authHandler.GetMe)
		protected.GET("/api/auth/me", authHandler.GetMe)
	}

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Auth Service HTTP server listening on port %s...", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Auth Service HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Auth Service...")

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}
	log.Println("Auth Service gracefully stopped.")
}
