package scheduler

import (
	"context"
	"fmt"
	"time"
)

func (a *App) cleanOldEvents(ctx context.Context) error {
	threshold := time.Now().AddDate(0, 0, -a.config.CleanThresholdDays)
	a.logger.Info(fmt.Sprintf("removing events older than %s", threshold))

	deletedCount, err := a.storage.DeleteEventsOlderThan(ctx, threshold)
	if err != nil {
		return fmt.Errorf("failed to clean old events: %w", err)
	}

	if deletedCount > 0 {
		a.logger.Info(fmt.Sprintf("removed %d old events", deletedCount))
	} else {
		a.logger.Info("no old events to remove")
	}
	return nil
}
