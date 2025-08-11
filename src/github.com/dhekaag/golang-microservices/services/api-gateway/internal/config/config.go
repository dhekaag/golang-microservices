package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Services  ServicesConfig
	RateLimit RateLimitConfig
	Session   SessionConfig
}

type ServerConfig struct {
	Port           string
	RequestTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type ServicesConfig struct {
	UserService    string
	ProductService string
	OrderService   string
}

type RateLimitConfig struct {
	RequestsPerMinute int
	WindowSize        time.Duration
}

type SessionConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	SessionTTL    time.Duration
	SessionPrefix string
}

func Load() *Config {

	return &Config{
		Server: ServerConfig{
			Port:           getEnv("PORT", "8080"),
			RequestTimeout: getDurationEnv("REQUEST_TIMEOUT", 30*time.Second),
			ReadTimeout:    getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:   getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
		},
		Services: ServicesConfig{
			UserService:    getEnv("USER_SERVICE_URL", "http://localhost:8081"),
			ProductService: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
			OrderService:   getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getIntEnv("RATE_LIMIT_RPM", 60),
			WindowSize:        getDurationEnv("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
		Session: SessionConfig{
			RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
			RedisPassword: getEnv("REDIS_PASSWORD", ""),
			RedisDB:       getIntEnv("REDIS_DB", 0),
			SessionTTL:    getDurationEnv("SESSION_TTL", 24*time.Hour),
			SessionPrefix: getEnv("SESSION_PREFIX", "session"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
