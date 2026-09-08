package main

import (
	"database/sql"
	"log"
	"time"

	"product-service/config"
	"product-service/handler"
	"product-service/repository"
	"product-service/service"

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
	query := `
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

		INSERT INTO products (name, description, price, stock, category) VALUES
		('ProBook Laptop 15"', 'High-performance laptop with 16GB RAM and 512GB SSD', 1299.99, 25, 'Electronics'),
		('UltraPhone 14 Pro', 'Flagship smartphone with triple camera system and OLED display', 999.00, 40, 'Electronics'),
		('NoiseCancelling Headphones', 'Wireless over-ear headphones with active noise cancellation', 249.50, 60, 'Audio'),
		('Mechanical RGB Keyboard', 'Tactile mechanical keyboard with customizable RGB backlighting', 89.99, 100, 'Accessories'),
		('Wireless Ergonomic Mouse', 'Precision optical mouse with multi-device bluetooth pairing', 49.99, 120, 'Accessories'),
		('4K UltraHD Monitor 27"', 'IPS display panel with 144Hz refresh rate and HDR400', 399.99, 15, 'Electronics'),
		('SmartWatch Pro V2', 'Fitness tracking smartwatch with heart rate & ECG sensors', 199.95, 50, 'Wearables'),
		('OctaTab 11 Tablet', '11-inch tablet with stylus support and all-day battery life', 499.00, 30, 'Electronics'),
		('Mirrorless 4K Camera', 'Professional digital camera with 24MP sensor and 4K video recording', 849.99, 10, 'Photography'),
		('NextGen Gaming Console 1TB', 'Ultra-fast SSD gaming console supporting up to 120 FPS output', 499.99, 20, 'Gaming')
		ON CONFLICT DO NOTHING;
	`
	_, err := db.Exec(query)
	return err
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

	router := gin.Default()

	// Routes
	router.GET("/health", h.HealthCheck)
	router.GET("/products/health", h.HealthCheck)
	router.GET("/products", h.ListProducts)
	router.GET("/products/:id", h.GetProduct)
	router.POST("/products", h.CreateProduct)
	router.PUT("/products/:id", h.UpdateProduct)
	router.DELETE("/products/:id", h.DeleteProduct)

	log.Printf("Product Service starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Product Service failed to start: %v", err)
	}
}
