package config

import (
	"os"
	"strconv"
	"time"

	"github.com/dhekaag/golang-microservices/shared/pkg/database"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database *database.DatabaseConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		println("Warning: Error loading .env file:", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8081"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
		},
		Database: &database.DatabaseConfig{
			HOST:            getEnv("DB_HOST", "localhost"),
			Port:            getIntEnv("DB_PORT", 3306),
			USER:            getEnv("DB_USER", "root"),
			PASSWORD:        getEnv("DB_PASSWORD", ""),
			DBNAME:          getEnv("DB_NAME", "microservice_users"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 25),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 200),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
