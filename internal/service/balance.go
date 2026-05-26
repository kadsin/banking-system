package service

import (
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
)

func NewBalanceService(ledger contracts.MainLedgerRepository) *balanceService {
	return &balanceService{
		ledger: ledger,
	}
}

type balanceService struct {
	ledger contracts.MainLedgerRepository
}

func (s *balanceService) Get(accountID string) (int64, error) {
	return s.ledger.Get(accountID)
}

func (s *balanceService) Adjust(accountID string, delta int64) error {
	return s.ledger.Adjust(accountID, delta)
}

func (s *balanceService) Refund(transaction domain.Transaction) error {
	if err := s.ledger.Adjust(transaction.FromAccountID, transaction.Amount); err != nil {
		return err
	}

	return s.ledger.Adjust(transaction.ToAccountID, -transaction.Amount)
}
