package model

import (
	"context"
	"fmt"
	"js-centralized-wallet/pkg/trace"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (m *Model) TransferBalanceV2(ctx context.Context, sourceUserId, destUserId uint64, amount int64) error {
	ctx, lg := trace.Logger(ctx)

	lg.Info(fmt.Sprintf("Starts transferring $%d from user_id %d to user_id %d", amount, sourceUserId, destUserId))

	// Simulate slow processing / delay
	time.Sleep(3 * time.Second)

	err := m.db.Transaction(func(tx *gorm.DB) error {

		// Lock wallets
		sourceWallet, destWallet, err := LockWalletsBalanceByUserId(ctx, sourceUserId, destUserId, tx)
		if err != nil {
			return err
		}
		if sourceWallet.Balance < amount {

			// Might have issue even though checked before pushing into channel
			// TODO:
			// 1) Add retry mechanism in the future
			// OR
			// 2) Push into persistent storage to notify users
			return ErrBalanceInsufficient
		}
		sourceWallet.Balance -= amount
		destWallet.Balance += amount

		if err := tx.Save(sourceWallet).Error; err != nil {
			return err
		}
		if err := tx.Save(destWallet).Error; err != nil {
			return err
		}

		// When we get listing / sync, we filter by DestWalletId with the amount
		sourceTransaction := Transaction{
			TransactionUUID: uuid.New().String(),
			SourceWalletId:  destWallet.Id,
			DestWalletId:    sourceWallet.Id,
			Amount:          amount * -1,
			Type:            TRANSACTION_TYPE_TRANSFER,
		}

		// When we get listing / sync, we filter by DestWalletId with the amount
		destTransaction := Transaction{
			TransactionUUID: uuid.New().String(),
			SourceWalletId:  sourceWallet.Id,
			DestWalletId:    destWallet.Id,
			Amount:          amount,
			Type:            TRANSACTION_TYPE_TRANSFER,
		}

		if err := tx.Create(&sourceTransaction).Error; err != nil {
			return err
		}
		if err := tx.Create(&destTransaction).Error; err != nil {
			return err
		}

		return err
	})

	return err
}
