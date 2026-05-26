package datalayer

import (
	"errors"

	"github.com/kadsin/banking-system/internal/cache"
)

func NewTxIdempotencyRepository(redis *cache.Cache) *TxIdempotencyRepository {
	return &TxIdempotencyRepository{
		cache: redis,
	}
}

type TxIdempotencyRepository struct {
	cache *cache.Cache
}

func (r *TxIdempotencyRepository) Get(idempotencyKey string) (string, bool, error) {
	txID, err := r.cache.Get(idempotencyKey)
	if err != nil {
		if errors.Is(err, cache.ErrKeyNotFound) {
			return "", false, nil
		}

		return "", false, err
	}

	return txID, true, nil
}

func (r *TxIdempotencyRepository) Set(idempotencyKey string, txID string) error {
	return r.cache.Set(idempotencyKey, txID)
}
