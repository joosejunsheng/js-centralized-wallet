package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

func (s *Server) StartScheduler() error {
	c := cron.New()

	// Run at every minute
	_, err := c.AddFunc("0 0 * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := s.model.SyncWalletSnapshots(ctx)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to take wallet snapshot: %v", err))
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	c.Start()

	return nil
}
