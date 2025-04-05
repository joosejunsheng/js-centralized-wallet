package model

import (
	"context"
	"fmt"
	"js-centralized-wallet/internal/constants"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (m *Model) connectRedis() error {

	options := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", constants.REDIS_HOST, constants.REDIS_PORT),
		DB:   0,
	}

	client := redis.NewClient(options)

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	m.redis = client

	slog.Info("connected to Redis")

	return nil
}

func (m *Model) connectDB() error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		constants.POSTGRES_USER,
		constants.POSTGRES_PASSWORD,
		constants.POSTGRES_HOST,
		constants.POSTGRES_DB,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	m.db = db

	slog.Info("connected to database")

	return nil
}

func (m *Model) migrate() error {
	slog.Info("Running GORM AutoMigrate")

	err := m.db.AutoMigrate(&User{}, &Wallet{}, &Transaction{})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database migrated successfully")

	if err := m.seed(); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	return nil
}

func (m *Model) seed() error {
	slog.Info("Seeding database")

	var userCount int64
	m.db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		users := []User{
			{Name: "User A", Email: "user_a@crypto.com"},
			{Name: "User B", Email: "user_a@crypto.com"},
		}

		if err := m.db.Create(&users).Error; err != nil {
			return fmt.Errorf("failed to seed users: %w", err)
		}
		slog.Info("Users seeded successfully")
	}

	var walletCount int64
	m.db.Model(&Wallet{}).Count(&walletCount)
	if walletCount == 0 {
		wallets := []Wallet{
			{UserId: 1, Balance: 1_000_000_000_000},
			{UserId: 2, Balance: 1_000_000_000_000},
		}

		if err := m.db.Create(&wallets).Error; err != nil {
			return fmt.Errorf("failed to seed wallets: %w", err)
		}
		slog.Info("Wallets seeded successfully")
	}

	slog.Info("Database seeding completed")
	return nil
}
