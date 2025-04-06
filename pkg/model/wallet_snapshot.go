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

// TODO:
// Adds up previous day transactions logs, and adds last wallet snapshot record from the day before, summing up a total to compare to balance
// If does not tally, update the wallet balance to the sum amount
func (m *Model) SyncWalletSnapshots(ctx context.Context) error {
	// ctx, _ = trace.Logger(ctx)

	return nil
}
