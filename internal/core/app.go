package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/kadsin/banking-system/internal/queue"
)

func New(txRepo contracts.MainTransactionRepository, balanceRepo contracts.MainLedgerRepository, q *queue.Queue) *App {
	return &App{
		txRepo:      txRepo,
		balanceRepo: balanceRepo,
		q:           q,
	}
}

type App struct {
	txRepo      contracts.MainTransactionRepository
	balanceRepo contracts.MainLedgerRepository
	q           *queue.Queue
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
			if _, err := a.PullAndProcess(batchSize); err != nil {
				return err
			}
		}
	}
}

func (a *App) PullAndProcess(limit int) (int, error) {
	messages, err := a.q.Fetch(config.Env.Topics.Transactions, limit)
	if err != nil {
		return 0, err
	}

	if len(messages) == 0 {
		return 0, nil
	}

	transactions := make([]domain.Transaction, 0, len(messages))
	for _, msg := range messages {
		var tx domain.Transaction
		if err := json.Unmarshal(msg.Value, &tx); err != nil {
			return 0, err
		}

		transactions = append(transactions, tx)
	}

	if err := a.txRepo.BulkCreate(transactions); err != nil {
		return 0, err
	}

	for i, tx := range transactions {
		if err := a.balanceRepo.Adjust(tx.FromAccountID, -tx.Amount); err != nil {
			transactions[i].Status = domain.TransactionStatusFailed

			if err := a.publishFailed(transactions[i]); err != nil {
				return 0, err
			}
			continue
		}

		if err := a.balanceRepo.Adjust(tx.ToAccountID, tx.Amount); err != nil {
			transactions[i].Status = domain.TransactionStatusFailed

			if err := a.publishFailed(transactions[i]); err != nil {
				return 0, err
			}
			continue
		}
	}

	lastOffset := messages[len(messages)-1].Offset
	if err := a.q.Commit(config.Env.Topics.Transactions, lastOffset); err != nil {
		return 0, err
	}

	return len(transactions), nil
}

func (a *App) publishFailed(tx domain.Transaction) error {
	payload, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	_, err = a.q.Publish(config.Env.Topics.Failed, payload)
	return err
}
