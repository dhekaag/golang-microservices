package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/config"
	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/handler"
	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/proxy"
	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/router"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
	cfg := config.Load()
	bootstrap, err := config.BootStrap(cfg)
	if err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}
	defer bootstrap.Cleanup()
	appLogger := bootstrap.Log

	serviceProxy := proxy.NewServiceProxy(&cfg.Services)
	appLogger.InfoMsg("Service proxy initialized",
		"user_service", cfg.Services.UserService,
		"product_service", cfg.Services.ProductService,
		"order_service", cfg.Services.OrderService,
	)

	authHandler := handler.NewAuthHandler(&cfg.Services, bootstrap.SessionManager)
	apiRouter := router.NewRouter(serviceProxy, authHandler, cfg)

	appLogger.InfoMsg("API Gateway initialization completed")

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      apiRouter.SetupRoutes(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.InfoMsg("Starting HTTP server",
			"address", server.Addr,
			"read_timeout", cfg.Server.ReadTimeout,
			"write_timeout", cfg.Server.WriteTimeout,
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.ErrorMsg("❌ Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Log successful startup with connected services
	services := []string{
		cfg.Services.UserService,
		cfg.Services.ProductService,
		cfg.Services.OrderService,
	}
	logger.ServiceStarted(cfg.Server.Port, services...)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.InfoMsg("🔄 Shutting down API Gateway...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		appLogger.ErrorMsg("❌ Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.ServiceStopped()
}
