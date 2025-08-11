package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	HOST            string        `json:"host"`
	Port            int           `json:"port"`
	USER            string        `json:"user"`
	PASSWORD        string        `json:"password"`
	DBNAME          string        `json:"dbname"`
	SSLMode         string        `json:"ssl_mode"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

func NewDatabaseConnection(config DatabaseConfig) (*gorm.DB, error) {
	// Build optimized DSN for MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s&interpolateParams=true",
		config.USER,
		config.PASSWORD,
		config.HOST,
		config.Port,
		config.DBNAME,
	)

	gormConfig := &gorm.Config{
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	maxIdleConns := config.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 25
	}

	maxOpenConns := config.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 200
	}

	connMaxLifetime := config.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = 30 * time.Minute
	}

	connMaxIdleTime := config.ConnMaxIdleTime
	if connMaxIdleTime == 0 {
		connMaxIdleTime = 5 * time.Minute
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// Test the connection with timeout
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}
