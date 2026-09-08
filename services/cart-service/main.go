package main

import (
	"context"
	"log"
	"time"

	"cart-service/config"
	"cart-service/handler"
	"cart-service/repository"
	"cart-service/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

	router := gin.Default()

	// Routes
	router.GET("/health", h.HealthCheck)
	router.GET("/cart/health", h.HealthCheck)
	router.GET("/cart", h.GetCart)
	router.POST("/cart/items", h.AddToCart)
	router.PUT("/cart/items/:productId", h.UpdateItemQuantity)
	router.DELETE("/cart/items/:productId", h.RemoveFromCart)
	router.DELETE("/cart", h.ClearCart)

	log.Printf("Cart Service starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Cart Service failed to start: %v", err)
	}
}
