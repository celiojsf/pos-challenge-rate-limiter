package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis          RedisConfig
	RateLimitIP    int
	RateLimitToken int
	BlockDuration  int
	TokenLimits    map[string]int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	rateLimitIP, err := strconv.Atoi(getEnv("RATE_LIMIT_IP", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_IP: %w", err)
	}

	rateLimitToken, err := strconv.Atoi(getEnv("RATE_LIMIT_TOKEN", "100"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_TOKEN: %w", err)
	}

	blockDuration, err := strconv.Atoi(getEnv("BLOCK_DURATION_SECONDS", "300"))
	if err != nil {
		return nil, fmt.Errorf("invalid BLOCK_DURATION_SECONDS: %w", err)
	}

	config := &Config{
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		RateLimitIP:    rateLimitIP,
		RateLimitToken: rateLimitToken,
		BlockDuration:  blockDuration,
		TokenLimits:    make(map[string]int),
	}

	// Load custom token limits
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TOKEN_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				token := strings.TrimPrefix(parts[0], "TOKEN_")
				limit, err := strconv.Atoi(parts[1])
				if err == nil {
					config.TokenLimits[token] = limit
				}
			}
		}
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
