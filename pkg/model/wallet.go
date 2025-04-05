package model

import (
	"context"
	"fmt"
	"js-centralized-wallet/pkg/trace"
)

type Wallet struct {
	Base
	UserId  uint64 `json:"user_id"`
	Balance int64  `json:"balance"`
}

func (*Wallet) TableName() string {
	return "wallets"
}

func (m *Model) GetWalletBalance(ctx context.Context, userId uint64) (int64, error) {
	ctx, lg := trace.Logger(ctx)

	var balance int64

	if err := m.db.WithContext(ctx).
		Model(&Wallet{}).
		Select("balance").
		Where("user_id = ?", userId).
		Scan(&balance).Error; err != nil {
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	lg.Info(fmt.Sprintf("retrieved wallet balance for user = %d: $%d", userId, balance))

	return balance, nil
}
