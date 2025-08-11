package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dhekaag/golang-microservices/services/user-service/internal/config"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Bootstrap application
	bootstrap, err := config.Bootstrap(cfg)
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}
	defer bootstrap.Cleanup()

	appLogger := bootstrap.Logger
	appLogger.InfoMsg("User service initialization completed")

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      bootstrap.Router.SetupRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.InfoMsg("Starting HTTP server",
			"address", server.Addr,
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.ErrorMsg("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Log successful startup
	logger.ServiceStarted(cfg.Server.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.InfoMsg("Shutting down User service...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		appLogger.ErrorMsg("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.ServiceStopped()
}
