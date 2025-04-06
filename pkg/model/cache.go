package model

import (
	"context"
	"fmt"
	"js-centralized-wallet/pkg/trace"
)

func (m *Model) InvalidateWalletCache(ctx context.Context, userIds ...uint64) {

	ctx, lg := trace.Logger(ctx)

	for _, userId := range userIds {
		key := fmt.Sprintf("balance:%d", userId)
		if err := m.GetRedis().Del(ctx, key).Err(); err != nil {
			lg.Info(fmt.Sprintf("Failed to invalidate user %d balance cache: %v", userId, err))
		}

		key = fmt.Sprintf("transaction_history:%d-0-1-30", userId)
		if err := m.GetRedis().Del(ctx, key).Err(); err != nil {
			lg.Info(fmt.Sprintf("Failed to invalidate user %d history cache: %v", userId, err))
		}
	}
}
