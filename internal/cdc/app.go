package cdc

import (
	"context"
	"time"

	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/queue"
)

func New(outbox contracts.OutboxRepository, q *queue.Queue) *App {
	return &App{
		outbox: outbox,
		q:      q,
	}
}

type App struct {
	outbox contracts.OutboxRepository
	q      *queue.Queue
}

func (a *App) Run(ctx context.Context) error {
	pollInterval := 250 * time.Millisecond
	batchSize := 100

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := a.PullAndPush(batchSize); err != nil {
				return err
			}
		}
	}
}

func (a *App) PullAndPush(limit int) (int, error) {
	events, err := a.outbox.ListPending(limit)
	if err != nil {
		return 0, err
	}

	published := 0
	for _, event := range events {
		if _, err := a.q.Publish(event.Topic, event.Payload); err != nil {
			return published, err
		}

		if err := a.outbox.MarkProcessed(event.ID); err != nil {
			return published, err
		}

		published++
	}

	return published, nil
}
