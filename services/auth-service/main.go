package main

import (
	"database/sql"
	"log"
	"time"

	"auth-service/config"
	"auth-service/handler"
	"auth-service/middleware"
	"auth-service/repository"
	"auth-service/service"

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

	router := gin.Default()

	// Routes
	router.GET("/health", authHandler.HealthCheck)
	router.GET("/auth/health", authHandler.HealthCheck)
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)

	// Protected route
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(authSvc))
	{
		protected.GET("/me", authHandler.GetMe)
		protected.GET("/auth/me", authHandler.GetMe)
	}

	log.Printf("Auth Service starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Auth Service failed to start: %v", err)
	}
}
