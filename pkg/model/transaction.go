package model

import (
	"context"
	"fmt"
	"js-centralized-wallet/pkg/trace"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionType int

const (
	TRANSACTION_TYPE_DEPOSIT TransactionType = iota + 1
	TRANSACTION_TYPE_WITHDRAW
	TRANSACTION_TYPE_TRANSFER
)

func (t TransactionType) String() string {
	switch t {
	case TRANSACTION_TYPE_DEPOSIT:
		return "Deposit"
	case TRANSACTION_TYPE_WITHDRAW:
		return "Withdraw"
	case TRANSACTION_TYPE_TRANSFER:
		return "Transfer"
	default:
		return "-"
	}
}

type Transaction struct {
	Base
	TransactionUUID string          `json:"transaction_uuid"`
	SourceWalletId  uint64          `json:"source_wallet_id"`
	DestWalletId    uint64          `gorm:"index" json:"dest_wallet_id"`
	Amount          int64           `json:"amount"`
	Type            TransactionType `json:"type"`
}

func (*Transaction) TableName() string {
	return "transactions"
}

func (m *Model) GetTransactionHistory(ctx context.Context, userId uint64, transctionType int, pageInfo PageInfo) ([]Transaction, error) {

	var transactions []Transaction
	var walletId uint64
	var err error

	if err = m.db.
		Model(&Wallet{}).
		Select("id").
		Where("user_id = ?", userId).
		Scan(&walletId).Error; err != nil {
		return transactions, err
	}

	// query := m.db.Table("transactions").
	// 	Select(`transactions.*,
	//             u.email AS user_email`).
	// 	Joins(`JOIN wallets AS w ON transactions.source_wallet_id = w.id`).
	// 	Joins(`JOIN users AS u ON w.user_id = u.id`).
	// 	Where("transactions.dest_wallet_id = ?", walletId)

	query := m.db.Where("dest_wallet_id = ?", walletId)

	if transctionType > 0 && transctionType <= 3 {
		query = query.Where("type = ?", transctionType)
	}

	if pageInfo.Page < 1 {
		pageInfo.Page = 1
	}

	if pageInfo.PageSize == 0 || pageInfo.PageSize > 100 {
		pageInfo.PageSize = 30
	}

	offset := (pageInfo.Page - 1) * pageInfo.PageSize
	query = query.Offset(offset).Limit(pageInfo.PageSize).Order("created_at desc")

	if err = query.Find(&transactions).Error; err != nil {
		return transactions, err
	}

	return transactions, err

}

func (m *Model) Deposit(ctx context.Context, userId uint64, amount int64) (int64, error) {

	if amount < 1 {
		return 0, ErrInvalidAmount
	}

	var userWallet Wallet

	err := m.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id", userId).First(&userWallet).Error
		if err != nil {
			return err
		}

		userWallet.Balance += amount

		if err := tx.Save(userWallet).Error; err != nil {
			return err
		}

		transaction := Transaction{
			TransactionUUID: uuid.New().String(),
			SourceWalletId:  userWallet.Id,
			DestWalletId:    userWallet.Id,
			Amount:          amount,
			Type:            TRANSACTION_TYPE_DEPOSIT,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return err
	})

	return userWallet.Balance, err
}

func (m *Model) Withdraw(ctx context.Context, userId uint64, amount int64) (int64, error) {

	var userWallet Wallet

	err := m.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id", userId).First(&userWallet).Error
		if err != nil {
			return err
		}

		if userWallet.Balance < amount {
			return ErrBalanceInsufficient
		}

		userWallet.Balance -= amount

		if err := tx.Save(userWallet).Error; err != nil {
			return err
		}

		transaction := Transaction{
			TransactionUUID: uuid.New().String(),
			SourceWalletId:  userWallet.Id,
			DestWalletId:    userWallet.Id,
			Amount:          amount * -1,
			Type:            TRANSACTION_TYPE_WITHDRAW,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return err
	})

	return userWallet.Balance, err
}

func (m *Model) TransferBalance(ctx context.Context, sourceUserId, destUserId uint64, amount int64) error {
	ctx, lg := trace.Logger(ctx)

	lg.Info(fmt.Sprintf("Starts transferring $%d from user_id %d to user_id %d", amount, sourceUserId, destUserId))

	// Simulate slow process / delay
	time.Sleep(2 * time.Second)
	err := m.db.Transaction(func(tx *gorm.DB) error {

		// Lock wallets
		sourceWallet, destWallet, err := LockWalletsBalanceByUserId(ctx, sourceUserId, destUserId, tx)
		if err != nil {
			return err
		}
		if sourceWallet.Balance < amount {

			// V2 TO TAKE NOTE
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

// Always row lock smaller userId first to prevent deadlock
// Uses "UPDATE" lock instead of "SHARE" lock, stricter
func LockWalletsBalanceByUserId(c context.Context, sourceUserId, destUserId uint64, tx *gorm.DB) (Wallet, Wallet, error) {
	var sourceWallet, destWallet Wallet
	if sourceUserId < destUserId {
		err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id", sourceUserId).First(&sourceWallet).Error
		if err != nil {
			return sourceWallet, destWallet, err
		}
		err = tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id", destUserId).First(&destWallet).Error
		if err != nil {
			return sourceWallet, destWallet, err
		}
	} else {
		err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id", destUserId).First(&destWallet).Error
		if err != nil {
			return sourceWallet, destWallet, err
		}
		err = tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("user_id", sourceUserId).First(&sourceWallet).Error
		if err != nil {
			return sourceWallet, destWallet, err
		}
	}
	return sourceWallet, destWallet, nil
}
