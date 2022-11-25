package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/logging/logrus"
	"gorm.io/plugin/opentelemetry/tracing"
	"prom/core/domain/repository"
)

func New(conn string) (repository.Connection, error) {
	// TODO change this to a custom logger
	logger := logger.New(
		logrus.NewWriter(),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Warn,
			Colorful:      false,
		},
	)
	db, err := gorm.Open(mysql.Open(conn), &gorm.Config{Logger: logger})
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to db: %w", err)
	}


	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;",
		"db",
	)

	tx := db.Exec(createSQL)
	if tx.Error != nil {
		return nil, fmt.Errorf("Cannot create database: %w", tx.Error)
	}

	// Init models
	db.AutoMigrate(&User{})

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("Cannot initialize tracing for gorm: %w", tx.Error)
	}

  // TODO add config for these options
  sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err != nil {
		return nil, fmt.Errorf("Cannot get sqldb: %w", err)
	}

	return db, nil
}
