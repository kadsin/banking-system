package datalayer

import (
	"errors"
	"sync"

	"github.com/kadsin/banking-system/internal/domain"
)

var ErrTransactionNotFound = errors.New("transaction not found")

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		transactions: map[string]domain.Transaction{},
	}
}

type TransactionRepository struct {
	mu           sync.RWMutex
	transactions map[string]domain.Transaction
}

func (r *TransactionRepository) Create(transaction domain.Transaction) (domain.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[transaction.ID] = transaction

	return transaction, nil
}

func (r *TransactionRepository) GetByID(id string) (domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transaction, ok := r.transactions[id]
	if !ok {
		return domain.Transaction{}, ErrTransactionNotFound
	}

	return transaction, nil
}

func (r *TransactionRepository) ListByAccountID(accountID string) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	history := make([]domain.Transaction, 0)
	for _, transaction := range r.transactions {
		if transaction.FromAccountID == accountID || transaction.ToAccountID == accountID {
			history = append(history, transaction)
		}
	}

	return history, nil
}

func (r *TransactionRepository) UpdateStatus(id string, status domain.TransactionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	transaction, ok := r.transactions[id]
	if !ok {
		return ErrTransactionNotFound
	}

	transaction.Status = status
	r.transactions[id] = transaction

	return nil
}
