package config

import "os"

type Config struct {
	Port              string
	AuthServiceURL    string
	ProductServiceURL string
	CartServiceURL    string
	OrderServiceURL   string
	PaymentServiceURL string
	JWTSecret         string
}

func LoadConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		CartServiceURL:    getEnv("CART_SERVICE_URL", "http://localhost:8083"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8084"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
		JWTSecret:         getEnv("JWT_SECRET", "super-secret-jwt-key"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
