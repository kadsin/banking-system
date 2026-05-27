package contracts

import "github.com/kadsin/banking-system/internal/domain"

type TransactionService interface {
	GetByID(id string) (domain.Transaction, error)
	History(accountID string) ([]domain.Transaction, error)
}
