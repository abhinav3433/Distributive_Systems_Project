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

	"product-service/config"
	productgrpc "product-service/grpc"
	"product-service/handler"
	"product-service/repository"
	"product-service/service"
	productv1 "proto/product/v1"

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
				log.Println("Successfully connected to PostgreSQL product_db")
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

func ensureTableAndSeedsExists(db *sql.DB) error {
	tableQuery := `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price NUMERIC(10, 2) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			category VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(tableQuery); err != nil {
		return err
	}

	// 1. Clean up historical duplicates if any exist (retain lowest ID per product name)
	cleanupQuery := `
		DELETE FROM products
		WHERE id NOT IN (
			SELECT MIN(id) FROM products GROUP BY name
		);
	`
	if _, err := db.Exec(cleanupQuery); err != nil {
		log.Printf("Notice on duplicate cleanup: %v", err)
	}

	// 2. Upsert canonical seed catalog using explicit ID conflict handling
	// This guarantees that restarting the service will NEVER insert duplicate records!
	seedsQuery := `
		INSERT INTO products (id, name, description, price, stock, category) VALUES
		(1, 'ProBook Laptop 15"', 'High-performance laptop with 16GB RAM and 512GB SSD', 1299.99, 25, 'Electronics'),
		(2, 'UltraPhone 14 Pro', 'Flagship smartphone with triple camera system and OLED display', 999.00, 40, 'Electronics'),
		(3, 'NoiseCancelling Headphones', 'Wireless over-ear headphones with active noise cancellation', 249.50, 60, 'Audio'),
		(4, 'Mechanical RGB Keyboard', 'Tactile mechanical keyboard with customizable RGB backlighting', 89.99, 100, 'Accessories'),
		(5, 'Wireless Ergonomic Mouse', 'Precision optical mouse with multi-device bluetooth pairing', 49.99, 120, 'Accessories'),
		(6, '4K UltraHD Monitor 27"', 'IPS display panel with 144Hz refresh rate and HDR400', 399.99, 15, 'Electronics'),
		(7, 'SmartWatch Pro V2', 'Fitness tracking smartwatch with heart rate & ECG sensors', 199.95, 50, 'Wearables'),
		(8, 'OctaTab 11 Tablet', '11-inch tablet with stylus support and all-day battery life', 499.00, 30, 'Electronics'),
		(9, 'Mirrorless 4K Camera', 'Professional digital camera with 24MP sensor and 4K video recording', 849.99, 10, 'Photography'),
		(10, 'NextGen Gaming Console 1TB', 'Ultra-fast SSD gaming console supporting up to 120 FPS output', 499.99, 20, 'Gaming'),
		(11, 'SlimBook Air 13"', 'Ultra-lightweight laptop with M-series silicon and 18-hour battery', 1099.00, 20, 'Electronics'),
		(12, 'DeskStation Mini PC', 'Compact desktop computer with 32GB RAM and dual 4K display output', 749.50, 18, 'Electronics'),
		(13, 'True Wireless Pro Earbuds', 'In-ear buds with spatial audio, transparency mode, and wireless charging case', 159.99, 85, 'Audio'),
		(14, 'Hi-Fi Studio Monitor Speakers', 'Pair of 5-inch powered desktop reference monitors for precision audio', 299.00, 22, 'Audio'),
		(15, 'Rugged Waterproof Bluetooth Speaker', '360-degree portable outdoor speaker with 24-hour battery and deep bass', 79.99, 70, 'Audio'),
		(16, '65W GaN USB-C Fast Charger', 'Ultra-compact 3-port wall charger for laptops, tablets, and phones', 39.99, 150, 'Accessories'),
		(17, 'Braided Thunderbolt 4 Cable 2m', '40Gbps high-speed data transfer and 100W power delivery cable', 29.50, 200, 'Accessories'),
		(18, 'Ergonomic Aluminum Laptop Stand', 'Adjustable height cooling stand crafted from anodized aerospace aluminum', 45.00, 65, 'Accessories'),
		(19, 'Active Fitness Tracker Band', 'Lightweight waterproof fitness band with sleep tracking and SpO2 monitor', 59.99, 90, 'Wearables'),
		(20, 'GPS Adventure Smartwatch', 'Rugged outdoor multi-sport watch with topographic maps and solar charging', 349.00, 25, 'Wearables'),
		(21, 'Professional Carbon Fiber Tripod', 'Lightweight 65-inch camera tripod with 360-degree fluid ball head', 139.99, 35, 'Photography'),
		(22, '50mm f/1.8 Prime Portrait Lens', 'Fast aperture portrait lens with beautiful bokeh and silent autofocus', 219.00, 28, 'Photography'),
		(23, 'Studio Bi-Color LED Ring Light', '18-inch dimmable ring light kit with phone holder and light stand', 69.99, 55, 'Photography'),
		(24, 'Wireless Pro Game Controller', 'Ergonomic wireless controller with hall-effect triggers and rear paddles', 69.99, 75, 'Gaming'),
		(25, '7.1 Surround Gaming Headset', 'Immersive headset with detachable noise-cancelling boom mic and memory foam', 89.99, 60, 'Gaming'),
		(26, 'RGB Extended Gaming Mousepad', 'Micro-textured waterproof cloth gaming mat with 14 RGB lighting modes', 24.99, 110, 'Gaming')
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			price = EXCLUDED.price,
			stock = EXCLUDED.stock,
			category = EXCLUDED.category,
			updated_at = NOW();
	`
	if _, err := db.Exec(seedsQuery); err != nil {
		return err
	}

	// 3. Advance sequence past highest seeded ID so new user-created products don't collide
	seqQuery := `SELECT setval('products_id_seq', (SELECT GREATEST(MAX(id), 1) FROM products));`
	if _, err := db.Exec(seqQuery); err != nil {
		log.Printf("Notice on setval: %v", err)
	}

	// 4. Create UNIQUE index on name to prevent any duplicate insertions
	uniqueIdxQuery := `CREATE UNIQUE INDEX IF NOT EXISTS idx_products_name ON products(name);`
	_, _ = db.Exec(uniqueIdxQuery)

	return nil
}

func main() {
	cfg := config.LoadConfig()

	db, err := initDB(cfg.GetDSN())
	var repo repository.ProductRepository

	if err != nil || db == nil {
		log.Printf("Warning: Database connection failed (%v). Falling back to mock product repository with seed data for standalone testing.", err)
		repo = repository.NewMockProductRepository()
	} else {
		if err := ensureTableAndSeedsExists(db); err != nil {
			log.Printf("Warning: Table/Seed migration failed: %v", err)
		}
		repo = repository.NewPostgresProductRepository(db)
	}

	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	// 1. Initialize and start gRPC Server
	grpcServer := googlegrpc.NewServer()
	productGRPCServer := productgrpc.NewProductGRPCServer(svc)
	productv1.RegisterProductServiceServer(grpcServer, productGRPCServer)

	// Register standard gRPC health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("product.v1.ProductService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("Product Service gRPC server listening on port %s...", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Product Service gRPC server error: %v", err)
		}
	}()

	// 2. HTTP Server (REST + Health)
	router := gin.Default()

	// Root endpoint (prevents 404 when tested directly)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "product-service",
			"status":    "UP",
			"http_port": cfg.Port,
			"grpc_port": cfg.GRPCPort,
			"endpoints": []string{"/health", "/products", "/products/:id"},
		})
	})

	// Routes
	router.GET("/health", h.HealthCheck)
	router.GET("/products/health", h.HealthCheck)
	router.GET("/api/products/health", h.HealthCheck)
	router.GET("/products", h.ListProducts)
	router.GET("/api/products", h.ListProducts)
	router.GET("/products/:id", h.GetProduct)
	router.GET("/api/products/:id", h.GetProduct)
	router.POST("/products", h.CreateProduct)
	router.POST("/api/products", h.CreateProduct)
	router.PUT("/products/:id", h.UpdateProduct)
	router.PUT("/api/products/:id", h.UpdateProduct)
	router.DELETE("/products/:id", h.DeleteProduct)
	router.DELETE("/api/products/:id", h.DeleteProduct)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Product Service HTTP server listening on port %s...", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Product Service HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Product Service...")

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}
	log.Println("Product Service gracefully stopped.")
}
