package config

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAddr string
	RedisPass string
	RedisDB   int
	Port      string
}

func LoadConfig() *Config {
	dbStr := getEnv("REDIS_DB", "0")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	return &Config{
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass: getEnv("REDIS_PASSWORD", ""),
		RedisDB:   db,
		Port:      getEnv("PORT", "8083"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
