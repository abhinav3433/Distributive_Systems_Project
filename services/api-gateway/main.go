package main

import (
	"log"

	"api-gateway/config"
	"api-gateway/handler"
	"api-gateway/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	router := gin.Default()

	// 1. Attach Global CORS Middleware
	router.Use(middleware.CORSMiddleware())

	// 2. Initialize Proxy Handler
	proxyHandler := handler.NewProxyHandler(
		cfg.AuthServiceURL,
		cfg.ProductServiceURL,
		cfg.CartServiceURL,
		cfg.OrderServiceURL,
		cfg.PaymentServiceURL,
	)

	// 3. Health Endpoint
	router.GET("/health", proxyHandler.HealthCheck)

	// 4. API Gateway Routes
	// Auth Service: /api/auth/* -> Auth Service
	router.Any("/api/auth", proxyHandler.AuthProxy())
	router.Any("/api/auth/*path", proxyHandler.AuthProxy())

	// Product Service: /api/products/* -> Product Service
	router.Any("/api/products", proxyHandler.ProductProxy())
	router.Any("/api/products/*path", proxyHandler.ProductProxy())

	// Cart Service: /api/cart/* -> Cart Service
	router.Any("/api/cart", proxyHandler.CartProxy())
	router.Any("/api/cart/*path", proxyHandler.CartProxy())

	// Order Service: /api/orders/* -> Order Service
	router.Any("/api/orders", proxyHandler.OrderProxy())
	router.Any("/api/orders/*path", proxyHandler.OrderProxy())

	// Payment Service: /api/payments/* -> Payment Service
	router.Any("/api/payments", proxyHandler.PaymentProxy())
	router.Any("/api/payments/*path", proxyHandler.PaymentProxy())

	log.Printf("==============================================")
	log.Printf("API Gateway listening on port %s", cfg.Port)
	log.Printf("Proxying:")
	log.Printf("  /api/auth/*     -> %s", cfg.AuthServiceURL)
	log.Printf("  /api/products/* -> %s", cfg.ProductServiceURL)
	log.Printf("  /api/cart/*     -> %s", cfg.CartServiceURL)
	log.Printf("  /api/orders/*   -> %s", cfg.OrderServiceURL)
	log.Printf("  /api/payments/* -> %s", cfg.PaymentServiceURL)
	log.Printf("==============================================")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
