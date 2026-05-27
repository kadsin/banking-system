package service

import (
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
)

func NewTransactionService(olap contracts.OlapRepository) *transactionService {
	return &transactionService{
		olap: olap,
	}
}

type transactionService struct {
	olap contracts.OlapRepository
}

func (s *transactionService) GetByID(id string) (domain.Transaction, error) {
	return s.olap.GetByID(id)
}

func (s *transactionService) History(accountID string) ([]domain.Transaction, error) {
	return s.olap.ListByAccountID(accountID)
}
