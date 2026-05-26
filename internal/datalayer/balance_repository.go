package datalayer

import (
	"errors"
	"strconv"

	"github.com/kadsin/banking-system/internal/cache"
)

var ErrInvalidBalanceValue = errors.New("invalid balance value")
var ErrInsufficientBalance = errors.New("insufficient balance")

func NewBalanceRepository(redis *cache.Cache) *BalanceRepository {
	return &BalanceRepository{cache: redis}
}

type BalanceRepository struct {
	cache *cache.Cache
}

func (r *BalanceRepository) Adjust(accountID string, delta int64) error {
	balance, err := r.Get(accountID)
	if err != nil {
		return err
	}

	balance = balance + delta
	if balance < 0 {
		return ErrInsufficientBalance
	}

	return r.cache.Set(accountID, strconv.FormatInt(balance, 10))
}

func (r *BalanceRepository) Get(accountID string) (int64, error) {
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
