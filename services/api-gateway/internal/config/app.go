package config

import (
	"context"
	"time"

	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/dhekaag/golang-microservices/shared/pkg/session"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

type BootstrapConfig struct {
	App            *Config
	Log            *logger.Logger
	Validate       *validator.Validate
	RedisClient    *redis.Client
	SessionManager *session.SessionManager
	// Remove handler and router from here to break the cycle
}

func BootStrap(config *Config) (*BootstrapConfig, error) {
	// Initialize logger
	loggerInstance, err := logger.Init(logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "api-gateway",
		Environment: "development",
	})
	if err != nil {
		return nil, err
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Session.RedisAddr,
		Password: config.Session.RedisPassword,
		DB:       config.Session.RedisDB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		loggerInstance.ErrorMsg("❌ Failed to connect to Redis", "error", err)
		return nil, err
	}

	sessionConfig := session.SessionConfig{
		RedisAddr:     config.Session.RedisAddr,
		RedisPassword: config.Session.RedisPassword,
		RedisDB:       config.Session.RedisDB,
		SessionTTL:    int(config.Session.SessionTTL.Seconds()),
		SessionPrefix: config.Session.SessionPrefix,
	}

	sessionManager, err := session.NewSessionManager(sessionConfig)
	if err != nil {
		loggerInstance.ErrorMsg("❌ Failed to initialize session manager", "error", err)
		return nil, err
	}

	// Initialize validator
	validator := validator.New()

	loggerInstance.InfoMsg("Core bootstrap completed successfully")

	return &BootstrapConfig{
		App:            config,
		Log:            loggerInstance,
		Validate:       validator,
		RedisClient:    redisClient,
		SessionManager: sessionManager,
	}, nil
}

// Cleanup method for graceful shutdown
func (bc *BootstrapConfig) Cleanup() error {

	// Close Redis client
	if bc.RedisClient != nil {
		if err := bc.RedisClient.Close(); err != nil {
			bc.Log.ErrorMsg("❌ Failed to close Redis connection", "error", err)
			return err
		}
		bc.Log.InfoMsg("Redis connection closed")
	}

	// Close session manager
	if bc.SessionManager != nil {
		if err := bc.SessionManager.Close(); err != nil {
			bc.Log.ErrorMsg("❌ Failed to close session manager", "error", err)
			return err
		}
		bc.Log.InfoMsg("Session manager closed")
	}

	return nil
}
