package model

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWalletBalance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &Model{
		db: db,
	}

	wallet := Wallet{
		UserId:  42,
		Balance: 1000,
	}
	err := db.Create(&wallet).Error
	assert.NoError(t, err)

	balance, err := model.GetWalletBalance(context.Background(), wallet.UserId)

	assert.NoError(t, err)
	assert.Equal(t, int64(1000), balance)
}
