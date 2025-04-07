package model

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDepositSuccess(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

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
		t.Fatalf("expected wallet balance 150, got %d", wallet.Balance)
	}

	if newBalance != 150 {
		t.Fatalf("expected new balance 150, got %d", newBalance)
	}

	var count int64
	db.Model(&Transaction{}).Where("dest_wallet_id = ? AND type = ?", wallet.Id, TRANSACTION_TYPE_DEPOSIT).Count(&count)

	if count != 1 {
		t.Fatalf("expected 1 transaction, found %d", count)
	}
}

func TestDepositFailedInvalidAmount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

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

	_, err := model.Deposit(context.Background(), user.Id, 0)
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}

	var wallet Wallet
	if err := db.First(&wallet, "user_id = ?", user.Id).Error; err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	var count int64
	db.Model(&Transaction{}).Where("dest_wallet_id = ? AND type = ?", wallet.Id, TRANSACTION_TYPE_DEPOSIT).Count(&count)

	if count > 0 {
		t.Fatalf("expected 0 transaction, found %d", count)
	}
}

func TestDepositFailedInvalidUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}
	user := User{}

	_, err := model.Deposit(context.Background(), user.Id, 50)
	if err == nil {
		t.Fatalf("expected invalid user, but got none")
	}
}

func TestWithdrawSuccess(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}
	user := User{
		Name:  "User B",
		Email: "user_b@crypto.com",
		Wallet: Wallet{
			Balance: 100,
		},
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, err := model.Withdraw(context.Background(), user.Id, 50)
	if errors.Is(err, ErrBalanceInsufficient) {
		t.Fatalf("expected ErrBalanceInsufficient, got %v", err)
	}

	var wallet Wallet
	if err := db.First(&wallet, "user_id = ?", user.Id).Error; err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	var count int64
	db.Model(&Transaction{}).Where("dest_wallet_id = ? AND type = ?", wallet.Id, TRANSACTION_TYPE_WITHDRAW).Count(&count)

	if count != 1 {
		t.Fatalf("expected 1 transaction, found %d", count)
	}
}

func TestWithdrawInsufficientBalance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}
	user := User{
		Name:  "User B",
		Email: "user_b@crypto.com",
		Wallet: Wallet{
			Balance: 50,
		},
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, err := model.Withdraw(context.Background(), user.Id, 100)
	if err == nil {
		t.Fatal("expected error due to insufficient balance")
	}
	if !errors.Is(err, ErrBalanceInsufficient) {
		t.Fatalf("expected ErrBalanceInsufficient, got %v", err)
	}

	var wallet Wallet
	if err := db.First(&wallet, "user_id = ?", user.Id).Error; err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	var count int64
	db.Model(&Transaction{}).Where("dest_wallet_id = ? AND type = ?", wallet.Id, TRANSACTION_TYPE_WITHDRAW).Count(&count)

	if count > 0 {
		t.Fatalf("expected 0 transaction, found %d", count)
	}
}

func TestWithdrawFailedInvalidUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}
	user := User{}

	_, err := model.Withdraw(context.Background(), user.Id, 50)
	if err == nil {
		t.Fatalf("expected invalid user, but got none")
	}
}

func TestTransferBalanceSuccess(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}

	source := User{
		Name:  "User A",
		Email: "User B@crypto.com",
	}
	dest := User{
		Name:  "User B",
		Email: "user_b@crypto.com",
	}
	if err := db.Create(&source).Error; err != nil {
		t.Fatalf("failed to create source user: %v", err)
	}
	if err := db.Create(&dest).Error; err != nil {
		t.Fatalf("failed to create dest user: %v", err)
	}

	sourceWallet := Wallet{
		UserId:  source.Id,
		Balance: 200,
	}
	destWallet := Wallet{
		UserId:  dest.Id,
		Balance: 0,
	}
	if err := db.Create(&sourceWallet).Error; err != nil {
		t.Fatalf("failed to create source wallet: %v", err)
	}
	if err := db.Create(&destWallet).Error; err != nil {
		t.Fatalf("failed to create dest wallet: %v", err)
	}

	err := model.TransferBalance(context.Background(), source.Id, dest.Id, 100)
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	if err := db.First(&sourceWallet, "user_id = ?", source.Id).Error; err != nil {
		t.Fatalf("failed to get source wallet: %v", err)
	}
	if err := db.First(&destWallet, "user_id = ?", dest.Id).Error; err != nil {
		t.Fatalf("failed to get dest wallet: %v", err)
	}

	if sourceWallet.Balance != 100 {
		t.Errorf("expected source wallet balance 100, got %d", sourceWallet.Balance)
	}
	if destWallet.Balance != 100 {
		t.Errorf("expected dest wallet balance 100, got %d", destWallet.Balance)
	}

	var transactions []Transaction
	if err := db.Where("type = ?", TRANSACTION_TYPE_TRANSFER).Find(&transactions).Error; err != nil {
		t.Fatalf("failed to query transactions: %v", err)
	}
	if len(transactions) != 2 {
		t.Errorf("expected 2 transfer transactions, got %d", len(transactions))
	}
}

func TestTransferBalanceInsufficientBalance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}

	source := User{
		Name:  "User A",
		Email: "user_a@crypto.com",
	}
	dest := User{
		Name:  "User B",
		Email: "user_b@crypto.com",
	}
	if err := db.Create(&source).Error; err != nil {
		t.Fatalf("failed to create source user: %v", err)
	}
	if err := db.Create(&dest).Error; err != nil {
		t.Fatalf("failed to create dest user: %v", err)
	}

	sourceWallet := Wallet{
		UserId:  source.Id,
		Balance: 50,
	}
	destWallet := Wallet{
		UserId:  dest.Id,
		Balance: 0,
	}
	if err := db.Create(&sourceWallet).Error; err != nil {
		t.Fatalf("failed to create source wallet: %v", err)
	}
	if err := db.Create(&destWallet).Error; err != nil {
		t.Fatalf("failed to create dest wallet: %v", err)
	}

	err := model.TransferBalance(context.Background(), source.Id, dest.Id, 100)

	if err != ErrBalanceInsufficient {
		t.Fatalf("expected ErrBalanceInsufficient, got %v", err)
	}

	if err := db.First(&sourceWallet, "user_id = ?", source.Id).Error; err != nil {
		t.Fatalf("failed to get source wallet: %v", err)
	}
	if err := db.First(&destWallet, "user_id = ?", dest.Id).Error; err != nil {
		t.Fatalf("failed to get dest wallet: %v", err)
	}

	if sourceWallet.Balance != 50 {
		t.Errorf("expected source wallet balance 50, got %d", sourceWallet.Balance)
	}
	if destWallet.Balance != 0 {
		t.Errorf("expected dest wallet balance 0, got %d", destWallet.Balance)
	}
}

func TestTransferBalanceInvalidUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}

	dest := User{
		Name:  "User B",
		Email: "user_b@crypto.com",
	}
	if err := db.Create(&dest).Error; err != nil {
		t.Fatalf("failed to create dest user: %v", err)
	}

	destWallet := Wallet{
		UserId:  dest.Id,
		Balance: 0,
	}
	if err := db.Create(&destWallet).Error; err != nil {
		t.Fatalf("failed to create dest wallet: %v", err)
	}

	invalidSourceUserId := uint64(9999)

	err := model.TransferBalance(context.Background(), invalidSourceUserId, dest.Id, 100)

	// Invalid User
	if err == nil {
		t.Fatalf("expected error, got %v", err)
	}

	if err := db.First(&destWallet, "user_id = ?", dest.Id).Error; err != nil {
		t.Fatalf("failed to get dest wallet: %v", err)
	}

	if destWallet.Balance != 0 {
		t.Errorf("expected dest wallet balance 0, got %d", destWallet.Balance)
	}

	var transactions []Transaction
	if err := db.Where("type = ?", TRANSACTION_TYPE_TRANSFER).Find(&transactions).Error; err != nil {
		t.Fatalf("failed to query transactions: %v", err)
	}
	if len(transactions) != 0 {
		t.Errorf("expected no transactions, got %d", len(transactions))
	}

	invalidDestUserId := uint64(9999)
	validDestUserId := dest.Id

	err = model.TransferBalance(context.Background(), validDestUserId, invalidDestUserId, 100)
	if err == nil {
		t.Fatalf("expected error for invalid destination user, got nil")
	}

	var sourceWallet Wallet
	if err := db.First(&sourceWallet, "user_id = ?", validDestUserId).Error; err != nil {
		t.Fatalf("failed to get source wallet: %v", err)
	}

	if sourceWallet.Balance != 0 {
		t.Errorf("expected source wallet balance to be 0, got %d", sourceWallet.Balance)
	}

	var transactionsAfterDestInvalid []Transaction
	if err := db.Where("type = ?", TRANSACTION_TYPE_TRANSFER).Find(&transactionsAfterDestInvalid).Error; err != nil {
		t.Fatalf("failed to query transactions: %v", err)
	}
	if len(transactionsAfterDestInvalid) != 0 {
		t.Errorf("expected no transactions for invalid destination, got %d", len(transactionsAfterDestInvalid))
	}
}

func TestGetTransactionHistory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}

	user := User{
		Name:  "User A",
		Email: "user_a@crypto.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	wallet := Wallet{
		UserId:  user.Id,
		Balance: 100,
	}
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	transactions := []Transaction{
		{DestWalletId: wallet.Id, Type: 1, Amount: 10},
		{DestWalletId: wallet.Id, Type: 2, Amount: 20},
		{DestWalletId: wallet.Id, Type: 3, Amount: 30},
	}

	for _, txn := range transactions {
		if err := db.Create(&txn).Error; err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}
	}

	pageInfo := PageInfo{
		Page:     1,
		PageSize: 2,
	}
	transactionsRes, err := model.GetTransactionHistory(context.Background(), user.Id, 1, pageInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transactionsRes) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactionsRes))
	}

	transactionsRes, err = model.GetTransactionHistory(context.Background(), user.Id, 2, pageInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transactionsRes) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactionsRes))
	}

	userNoTransactions := User{
		Name:  "User B",
		Email: "user_b@crypto.com",
	}
	if err := db.Create(&userNoTransactions).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	transactionsRes, err = model.GetTransactionHistory(context.Background(), userNoTransactions.Id, 0, pageInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transactionsRes) != 0 {
		t.Errorf("expected no transactions for user with no transactions, got %d", len(transactionsRes))
	}
}

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = db.AutoMigrate(&User{}, &Wallet{}, &Transaction{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	cleanup := func() {
		db.Exec("DELETE FROM transactions")
		db.Exec("DELETE FROM wallets")
		db.Exec("DELETE FROM users")
	}

	return db, cleanup
}
