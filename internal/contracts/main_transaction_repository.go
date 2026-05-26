package contracts

import "github.com/kadsin/banking-system/internal/domain"

type MainTransactionRepository interface {
	BulkCreate(transactions []domain.Transaction) error
	List() ([]domain.Transaction, error)
}
