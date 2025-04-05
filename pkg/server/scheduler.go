package server

import (
	"context"
	"fmt"
	"js-centralized-wallet/pkg/trace"
	"time"

	"github.com/robfig/cron/v3"
)

func (s *Server) StartScheduler() error {
	c := cron.New()
	_, lg := trace.Logger(context.Background())

	// Run at every minute
	_, err := c.AddFunc("* * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := s.model.SyncWalletSnapshots(ctx)
		if err != nil {
			lg.Error(fmt.Sprintf("failed to take wallet snapshot: %ww", err))
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	c.Start()

	return nil
}
