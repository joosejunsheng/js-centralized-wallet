package model

import (
	"context"
	"fmt"
	"log/slog"
)

type User struct {
	Base
	Name   string `json:"name"`
	Email  string `json:"email"`
	Wallet Wallet `json:"wallet"`
}

func (*User) TableName() string {
	return "users"
}

func (m *Model) GetAllUsers(ctx context.Context, pageInfo PageInfo) ([]*User, error) {
	var users []*User

	limit := pageInfo.PageSize
	offset := (pageInfo.Page - 1) * pageInfo.PageSize

	if err := m.db.WithContext(ctx).Preload("Wallet").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	slog.Info("Successfully get all users", "count", len(users))
	return users, nil
}
