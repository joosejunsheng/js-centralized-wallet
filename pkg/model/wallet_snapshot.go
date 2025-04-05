package model

import (
	"context"
)

type WalletAmountSnapshot struct {
	Base
	WalletId uint64 `gorm:"index"`
	Amount   int64  `json:"amount"`
}

func (*WalletAmountSnapshot) TableName() string {
	return "wallet_snapshots"
}

func (m *Model) SyncWalletSnapshots(ctx context.Context) error {
	// ctx, _ = trace.Logger(ctx)

	return nil
}
