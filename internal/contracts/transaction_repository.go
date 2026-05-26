package contracts

import "github.com/kadsin/banking-system/internal/domain"

type TransactionRepository interface {
	Create(transaction domain.Transaction) (domain.Transaction, error)
	GetByID(id string) (domain.Transaction, error)
	ListByAccountID(accountID string) ([]domain.Transaction, error)
	UpdateStatus(id string, status domain.TransactionStatus) error
}
