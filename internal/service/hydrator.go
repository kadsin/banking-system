package service

import (
	"encoding/json"
	"errors"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/kadsin/banking-system/internal/queue"
)

func NewHydratorService(ledger contracts.LedgerRepository, hydratorRepo contracts.HydratorRepository, q *queue.Queue) *hydratorService {
	return &hydratorService{
		ledger: ledger,
		repo:   hydratorRepo,
		q:      q,
	}
}

type hydratorService struct {
	ledger contracts.LedgerRepository
	repo   contracts.HydratorRepository
	q      *queue.Queue
}

func (s *hydratorService) Repopulate(accountID string) (int64, error) {
	snapshot, err := s.repo.GetSnapshot(accountID)
	if err != nil {
		if !errors.Is(err, cache.ErrKeyNotFound) {
			return 0, err
		}

		snapshot = contracts.HydratorSnapshot{
			Balance:     0,
			QueueOffset: -1,
		}
	}

	currentBalance, lastOffset, err := s.repopulateFromQueue(snapshot, accountID)
	if err != nil {
		return 0, err
	}

	if err := s.ledger.Set(accountID, currentBalance); err != nil {
		return 0, err
	}

	err = s.repo.SetSnapshot(accountID, contracts.HydratorSnapshot{
		Balance:     currentBalance,
		QueueOffset: lastOffset,
	})
	if err != nil {
		return 0, err
	}

	return currentBalance, nil
}

func (s *hydratorService) repopulateFromQueue(snapshot contracts.HydratorSnapshot, accountID string) (currentBalance int64, lastOffset int64, err error) {
	currentBalance = snapshot.Balance
	lastOffset = snapshot.QueueOffset

	nextOffset := lastOffset + 1
	for {
		messages, fetchErr := s.q.FetchFromOffset(config.Env.Topics.Transactions, nextOffset, 100)
		if fetchErr != nil {
			return 0, 0, fetchErr
		}

		if len(messages) == 0 {
			break
		}

		for _, message := range messages {
			var tx domain.Transaction
			if decodeErr := json.Unmarshal(message.Value, &tx); decodeErr != nil {
				return 0, 0, decodeErr
			}

			if tx.FromAccountID == accountID {
				currentBalance -= tx.Amount
			}

			if tx.ToAccountID == accountID {
				currentBalance += tx.Amount
			}

			lastOffset = message.Offset
		}

		nextOffset = lastOffset + 1
	}

	return currentBalance, lastOffset, nil
}
