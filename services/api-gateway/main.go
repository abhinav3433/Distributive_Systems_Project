package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/config"
	"api-gateway/grpcclient"
	"api-gateway/handler"
	"api-gateway/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Initialize gRPC Clients for downstream services
	grpcClients, err := grpcclient.NewClients(cfg)
	if err != nil {
		log.Printf("[GATEWAY] Warning initializing gRPC clients: %v", err)
	}
	defer grpcClients.Close()

	// 2. Initialize Handlers
	authHandler := handler.NewAuthGRPCHandler(grpcClients.AuthClient)
	productHandler := handler.NewProductGRPCHandler(grpcClients.ProductClient)
	cartHandler := handler.NewCartGRPCHandler(grpcClients.CartClient, grpcClients.AuthClient)
	orderHandler := handler.NewOrderGRPCHandler(grpcClients.OrderClient, grpcClients.AuthClient)
	paymentHandler := handler.NewPaymentGRPCHandler(grpcClients.PaymentClient)

	// Keep HTTP Proxy Handler for legacy or fallback paths
	proxyHandler := handler.NewProxyHandler(
		cfg.AuthServiceURL,
		cfg.ProductServiceURL,
		cfg.CartServiceURL,
		cfg.OrderServiceURL,
		cfg.PaymentServiceURL,
	)

	router := gin.Default()

	// 3. Attach Global CORS Middleware
	router.Use(middleware.CORSMiddleware())

	// 4. Root Endpoint (prevents 404 when opening gateway URL in browser)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":        "api-gateway",
			"status":         "UP",
			"communication":  "gRPC (Gateway -> Microservices) / REST (Client -> Gateway)",
			"downstream_rpc": grpcClients.GetHealthStatus(),
			"routes": gin.H{
				"auth":     "/api/auth/*",
				"products": "/api/products/*",
				"cart":     "/api/cart/*",
				"orders":   "/api/orders/*",
				"payments": "/api/payments/*",
				"health":   "/health",
			},
		})
	})

	// 5. Health Check Endpoint (includes downstream gRPC health probe statuses)
	router.GET("/health", func(c *gin.Context) {
		downstream := grpcClients.GetHealthStatus()
		c.JSON(http.StatusOK, gin.H{
			"service":    "api-gateway",
			"status":     "UP",
			"downstream": downstream,
		})
	})

	// 6. Direct gRPC Routes (Client -> REST Gateway -> gRPC Microservices)

	// Auth Service (gRPC)
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", authHandler.GetMe)
		// Fallback for health check
		authGroup.GET("/health", proxyHandler.AuthProxy())
	}

	// Product Service (gRPC)
	productGroup := router.Group("/api/products")
	{
		productGroup.GET("", productHandler.ListProducts)
		productGroup.GET("/", productHandler.ListProducts)
		productGroup.GET("/:id", productHandler.GetProduct)
		productGroup.POST("", productHandler.CreateProduct)
		productGroup.POST("/", productHandler.CreateProduct)
		productGroup.PUT("/:id", productHandler.UpdateProduct)
		productGroup.DELETE("/:id", productHandler.DeleteProduct)
	}

	// Cart Service (gRPC)
	cartGroup := router.Group("/api/cart")
	{
		cartGroup.GET("", cartHandler.GetCart)
		cartGroup.GET("/", cartHandler.GetCart)
		cartGroup.POST("/items", cartHandler.AddToCart)
		cartGroup.PUT("/items/:productId", cartHandler.UpdateItem)
		cartGroup.DELETE("/items/:productId", cartHandler.RemoveItem)
		cartGroup.DELETE("", cartHandler.ClearCart)
		cartGroup.DELETE("/", cartHandler.ClearCart)
	}

	// Order Service (gRPC)
	orderGroup := router.Group("/api/orders")
	{
		orderGroup.POST("", orderHandler.CreateOrder)
		orderGroup.POST("/", orderHandler.CreateOrder)
		orderGroup.GET("", orderHandler.GetOrders)
		orderGroup.GET("/", orderHandler.GetOrders)
		orderGroup.GET("/:id", orderHandler.GetOrderByID)
	}

	// Payment Service (gRPC)
	paymentGroup := router.Group("/api/payments")
	{
		paymentGroup.GET("/:orderId", paymentHandler.GetPaymentByOrderID)
		paymentGroup.POST("", paymentHandler.ProcessPayment)
		paymentGroup.POST("/", paymentHandler.ProcessPayment)
	}

	log.Printf("==========================================================")
	log.Printf(" API Gateway with gRPC Integration Online")
	log.Printf(" HTTP Port: %s", cfg.Port)
	log.Printf(" Downstream gRPC Targets:")
	log.Printf("   Auth Service:    %s", cfg.AuthGRPCAddr)
	log.Printf("   Product Service: %s", cfg.ProductGRPCAddr)
	log.Printf("   Cart Service:    %s", cfg.CartGRPCAddr)
	log.Printf("   Order Service:   %s", cfg.OrderGRPCAddr)
	log.Printf("   Payment Service: %s", cfg.PaymentGRPCAddr)
	log.Printf("==========================================================")

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API Gateway HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("API Gateway HTTP server forced to shutdown: %v", err)
	}
	log.Println("API Gateway gracefully stopped.")
}
