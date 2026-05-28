package config

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
	Port        string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://medium:***@localhost:5432/medium?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
