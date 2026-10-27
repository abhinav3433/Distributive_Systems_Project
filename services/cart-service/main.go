package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cart-service/config"
	cartgrpc "cart-service/grpc"
	"cart-service/handler"
	"cart-service/repository"
	"cart-service/service"
	cartv1 "proto/cart/v1"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func initRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	log.Printf("Successfully connected to Redis server at %s", cfg.RedisAddr)
	return client, nil
}

func main() {
	cfg := config.LoadConfig()

	redisClient, err := initRedis(cfg)
	var repo repository.CartRepository

	if err != nil || redisClient == nil {
		log.Printf("Warning: Redis connection failed (%v). Falling back to mock cart repository for standalone testing.", err)
		repo = repository.NewMockCartRepository()
	} else {
		repo = repository.NewRedisCartRepository(redisClient)
	}

	svc := service.NewCartService(repo)
	h := handler.NewCartHandler(svc)

	// 1. Initialize and start gRPC Server
	grpcServer := googlegrpc.NewServer()
	cartGRPCServer := cartgrpc.NewCartGRPCServer(svc)
	cartv1.RegisterCartServiceServer(grpcServer, cartGRPCServer)

	// Register standard gRPC health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("cart.v1.CartService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("Cart Service gRPC server listening on port %s...", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Cart Service gRPC server error: %v", err)
		}
	}()

	// 2. HTTP Server (REST + Health)
	router := gin.Default()

	// Root endpoint (prevents 404 when tested directly)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "cart-service",
			"status":    "UP",
			"http_port": cfg.Port,
			"grpc_port": cfg.GRPCPort,
			"endpoints": []string{"/health", "/cart", "/cart/items"},
		})
	})

	// Routes
	router.GET("/health", h.HealthCheck)
	router.GET("/cart/health", h.HealthCheck)
	router.GET("/api/cart/health", h.HealthCheck)
	router.GET("/cart", h.GetCart)
	router.GET("/api/cart", h.GetCart)
	router.POST("/cart/items", h.AddToCart)
	router.POST("/api/cart/items", h.AddToCart)
	router.PUT("/cart/items/:productId", h.UpdateItemQuantity)
	router.PUT("/api/cart/items/:productId", h.UpdateItemQuantity)
	router.DELETE("/cart/items/:productId", h.RemoveFromCart)
	router.DELETE("/api/cart/items/:productId", h.RemoveFromCart)
	router.DELETE("/cart", h.ClearCart)
	router.DELETE("/api/cart", h.ClearCart)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Cart Service HTTP server listening on port %s...", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Cart Service HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Cart Service...")

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}
	log.Println("Cart Service gracefully stopped.")
}
