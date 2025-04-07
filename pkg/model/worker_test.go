package model

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
)

type FakeTransferModel struct {
	db         *gorm.DB
	Transfers  []TransferJob
	CacheCalls []TransferJob
}

func (f *FakeTransferModel) TransferBalance(ctx context.Context, source, dest uint64, amount int64) error {
	f.Transfers = append(f.Transfers, TransferJob{ctx, source, dest, amount})
	return nil
}

func (f *FakeTransferModel) InvalidateWalletCache(ctx context.Context, userIds ...uint64) {
	f.CacheCalls = append(f.CacheCalls, TransferJob{ctx, userIds[0], userIds[1], 0})
}

func TestTransferWorkerPool(t *testing.T) {
	jobChan := make(chan TransferJob, 1)

	db, cleanup := setupTestDB(t)
	defer cleanup()

	fakeTransferModel := &FakeTransferModel{
		db: db,
	}

	pool := NewTransferWorkerPool(
		fakeTransferModel,
		jobChan,
		WithNumWorkers(1),
	)

	pool.Start()

	job := TransferJob{
		Ctx:          context.Background(),
		SourceUserId: 1,
		DestUserId:   2,
		Amount:       100,
	}
	jobChan <- job

	time.Sleep(100 * time.Millisecond)

	if len(fakeTransferModel.Transfers) != 1 {
		t.Fatalf("expected 1 transfer call, got %d", len(fakeTransferModel.Transfers))
	}
	if fakeTransferModel.Transfers[0].Amount != 100 {
		t.Errorf("expected amount 100, got %d", fakeTransferModel.Transfers[0].Amount)
	}

	if len(fakeTransferModel.CacheCalls) != 1 {
		t.Fatalf("expected 1 cache invalidation, got %d", len(fakeTransferModel.CacheCalls))
	}
}
