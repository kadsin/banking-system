package saga

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/kadsin/banking-system/internal/queue"
)

func New(balance contracts.BalanceService, q *queue.Queue) *App {
	return &App{
		balance: balance,
		q:       q,
	}
}

type App struct {
	balance contracts.BalanceService
	q       *queue.Queue
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
			if _, err := a.pullAndRefund(batchSize); err != nil {
				return err
			}
		}
	}
}

func (a *App) pullAndRefund(limit int) (int, error) {
	messages, err := a.q.Fetch(config.Env.Topics.Failed, limit)
	if err != nil {
		return 0, err
	}

	if len(messages) == 0 {
		return 0, nil
	}

	count := 0
	for _, msg := range messages {
		var tx domain.Transaction
		if err := json.Unmarshal(msg.Value, &tx); err != nil {
			return count, err
		}

		if err := a.balance.Refund(tx); err != nil {
			return count, err
		}
		count++
	}

	lastOffset := messages[len(messages)-1].Offset
	if err := a.q.Commit(config.Env.Topics.Failed, lastOffset); err != nil {
		return count, err
	}

	return count, nil
}
