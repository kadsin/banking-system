package datalayer

import (
	"errors"
	"strconv"

	"github.com/kadsin/banking-system/internal/cache"
)

var ErrInvalidBalanceValue = errors.New("invalid balance value")
var ErrInsufficientBalance = errors.New("insufficient balance")

func NewLedgerRepository(redis *cache.Cache) *LedgerRepository {
	return &LedgerRepository{cache: redis}
}

type LedgerRepository struct {
	cache *cache.Cache
}

func (r *LedgerRepository) Adjust(accountID string, delta int64) error {
	balance, err := r.Get(accountID)
	if err != nil {
		return err
	}

	balance = balance + delta
	if balance < 0 {
		return ErrInsufficientBalance
	}

	return r.Set(accountID, balance)
}

func (r *LedgerRepository) Get(accountID string) (int64, error) {
	value, err := r.cache.Get(accountID)
	if err != nil {
		return 0, err
	}

	balance, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, ErrInvalidBalanceValue
	}

	return balance, nil
}

func (r *LedgerRepository) Set(accountID string, balance int64) error {
	return r.cache.Set(accountID, strconv.FormatInt(balance, 10))
}
