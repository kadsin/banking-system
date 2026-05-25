package datalayer

import (
	"errors"
	"sync"

	"github.com/kadsin/banking-system/internal/domain"
)

var ErrAccountNotFound = errors.New("account not found")

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{
		accounts: map[string]domain.Account{},
	}
}

type AccountRepository struct {
	mu       sync.RWMutex
	accounts map[string]domain.Account
}

func (r *AccountRepository) Create(account domain.Account) (domain.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.accounts[account.ID] = account
	return account, nil
}

func (r *AccountRepository) GetByID(id string) (domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, ok := r.accounts[id]
	if !ok {
		return domain.Account{}, ErrAccountNotFound
	}

	return account, nil
}

func (r *AccountRepository) Update(account domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.accounts[account.ID]; !ok {
		return ErrAccountNotFound
	}

	r.accounts[account.ID] = account
	return nil
}
