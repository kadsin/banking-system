package datalayer

import (
	"sync"

	"github.com/kadsin/banking-system/internal/domain"
)

func NewMainTransactionRepository() *MainTransactionRepository {
	return &MainTransactionRepository{
		transactions: make([]domain.Transaction, 0),
	}
}

type MainTransactionRepository struct {
	mu           sync.RWMutex
	transactions []domain.Transaction
}

func (r *MainTransactionRepository) BulkCreate(transactions []domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions = append(r.transactions, transactions...)

	return nil
}

func (r *MainTransactionRepository) List() ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Transaction, 0, len(r.transactions))
	result = append(result, r.transactions...)

	return result, nil
}
