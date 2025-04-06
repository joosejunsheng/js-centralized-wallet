package model

import (
	"context"
	"fmt"
	"js-centralized-wallet/pkg/trace"
)

type TransferJob struct {
	Ctx          context.Context
	SourceUserId uint64
	DestUserId   uint64
	Amount       int64
}

type TransferWorkerPool struct {
	model      *Model
	jobChan    <-chan TransferJob
	numWorkers int
}

type TransferWorkerOption func(*TransferWorkerPool)

func NewTransferWorkerPool(model *Model, jobChan <-chan TransferJob, opts ...TransferWorkerOption) *TransferWorkerPool {
	pool := &TransferWorkerPool{
		model:      model,
		jobChan:    jobChan,
		numWorkers: 3,
	}
	for _, opt := range opts {
		opt(pool)
	}
	return pool
}

func WithNumWorkers(n int) TransferWorkerOption {
	return func(p *TransferWorkerPool) {
		p.numWorkers = n
	}
}

func (p *TransferWorkerPool) Start() {

	_, lg := trace.Logger(context.Background())

	for i := range p.numWorkers {
		go func(id int) {
			for job := range p.jobChan {
				err := p.model.TransferBalance(job.Ctx, job.SourceUserId, job.DestUserId, job.Amount)
				if err != nil {
					// TODO:
					// 1) Add retry mechanism in the future
					// OR
					// 2) Push into persistent storage to notify users
					lg.Info(fmt.Sprintf("[worker %d] transfer failed - FROM USER %d TO USER %d, AMOUNT %d ERR: %v", id, job.SourceUserId, job.DestUserId, job.Amount, err))
				} else {
					p.model.InvalidateWalletCache(job.Ctx, job.SourceUserId, job.DestUserId)
					lg.Info(fmt.Sprintf("[worker %d] transfer success - FROM USER %d TO USER %d, AMOUNT %d", id, job.SourceUserId, job.DestUserId, job.Amount))
				}
			}
		}(i)
	}
}
