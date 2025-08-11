package config

import (
	"github.com/dhekaag/golang-microservices/services/user-service/internal/handler"
	"github.com/dhekaag/golang-microservices/services/user-service/internal/repository"
	"github.com/dhekaag/golang-microservices/services/user-service/internal/router"
	"github.com/dhekaag/golang-microservices/services/user-service/internal/service"
	"github.com/dhekaag/golang-microservices/shared/pkg/database"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB          *gorm.DB
	Config      *Config
	Logger      *logger.Logger
	Validator   *validator.Validate
	UserRepo    repository.UserRepository
	UserService service.UserService
	UserHandler *handler.UserHandler
	Router      *router.Router
}

func Bootstrap(config *Config) (*BootstrapConfig, error) {
	// Initialize logger
	loggerInstance, err := logger.Init(logger.Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "user-service",
		Environment: "development",
	})
	if err != nil {
		return nil, err
	}

	loggerInstance.InfoMsg("Initializing user service...")

	// Initialize database
	loggerInstance.InfoMsg("Connecting to database...")
	db, err := database.NewDatabaseConnection(*config.Database)
	if err != nil {
		loggerInstance.ErrorMsg("Failed to connect to database", "error", err)
		return nil, err
	}
	loggerInstance.InfoMsg("Database connected successfully")

	// Initialize validator
	validator := validator.New()
	loggerInstance.InfoMsg("Validator initialized")

	// Initialize repository
	userRepo := repository.NewUserRepository(db)
	loggerInstance.InfoMsg("Repository initialized")

	// Initialize service
	userService := service.NewUserService(userRepo, loggerInstance)
	loggerInstance.InfoMsg("Service initialized")

	// Initialize handler
	userHandler := handler.NewUserHandler(userService, validator, loggerInstance)
	loggerInstance.InfoMsg("Handler initialized")

	// Initialize router
	userRouter := router.NewRouter(userHandler)
	loggerInstance.InfoMsg("Router initialized")

	loggerInstance.InfoMsg("User service bootstrap completed successfully")

	return &BootstrapConfig{
		DB:          db,
		Config:      config,
		Logger:      loggerInstance,
		Validator:   validator,
		UserRepo:    userRepo,
		UserService: userService,
		UserHandler: userHandler,
		Router:      userRouter,
	}, nil
}

func (bc *BootstrapConfig) Cleanup() error {
	bc.Logger.InfoMsg("🧹 Starting cleanup process...")

	if bc.DB != nil {
		bc.Logger.InfoMsg("Closing database connection...")
		sqlDB, err := bc.DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				bc.Logger.ErrorMsg("Failed to close database connection", "error", err)
				return err
			}
		}
		bc.Logger.InfoMsg("Database connection closed")
	}

	bc.Logger.InfoMsg("Cleanup completed successfully")
	return nil
}
