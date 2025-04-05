package model

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeposit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = db.AutoMigrate(&User{}, &Wallet{}, &Transaction{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	model := &Model{
		db: db,
	}
	user := User{
		Name:  "User A",
		Email: "user_a@crypto.com",
		Wallet: Wallet{
			Balance: 100,
		},
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	newBalance, err := model.Deposit(context.Background(), user.Id, 50)
	if err != nil {
		t.Fatalf("failed to deposit: %v", err)
	}

	var wallet Wallet
	if err := db.First(&wallet, "user_id = ?", user.Id).Error; err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if wallet.Balance != 150 {
		t.Errorf("expected wallet balance 150, got %d", wallet.Balance)
	}

	if newBalance != 150 {
		t.Errorf("expected new balance 150, got %d", newBalance)
	}

	var count int64
	db.Model(&Transaction{}).Where("source_wallet_id = ? AND type = ?", wallet.Id, TRANSACTION_TYPE_DEPOSIT).Count(&count)

	if count != 1 {
		t.Errorf("expected 1 transaction, found %d", count)
	}
}
