package service

import (
	"errors"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
)

func NewBalanceService(ledger contracts.LedgerRepository, hydrator contracts.HydratorService) *balanceService {
	return &balanceService{
		ledger:   ledger,
		hydrator: hydrator,
	}
}

type balanceService struct {
	ledger   contracts.LedgerRepository
	hydrator contracts.HydratorService
}

func (s *balanceService) Get(accountID string) (int64, error) {
	balance, err := s.ledger.Get(accountID)
	if err == nil {
		return balance, nil
	}

	if err := s.repopulateOnCacheMiss(accountID, err); err != nil {
		return 0, err
	}

	return s.ledger.Get(accountID)
}

func (s *balanceService) Adjust(accountID string, delta int64) error {
	err := s.ledger.Adjust(accountID, delta)
	if err == nil {
		return nil
	}

	if err := s.repopulateOnCacheMiss(accountID, err); err != nil {
		return err
	}

	return s.ledger.Adjust(accountID, delta)
}

func (s *balanceService) Refund(transaction domain.Transaction) error {
	if err := s.Adjust(transaction.FromAccountID, transaction.Amount); err != nil {
		return err
	}

	return s.Adjust(transaction.ToAccountID, -transaction.Amount)
}

func (s *balanceService) repopulateOnCacheMiss(accountID string, currentErr error) error {
	if !errors.Is(currentErr, cache.ErrKeyNotFound) {
		return currentErr
	}

	_, err := s.hydrator.Repopulate(accountID)
	return err
}
