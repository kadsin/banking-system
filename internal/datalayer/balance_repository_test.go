package datalayer

import (
	"testing"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestBalanceRepositorySetAndGet(t *testing.T) {
	redis := cache.New()
	repo := NewBalanceRepository(redis)

	err := redis.Set("acc-1", "0")
	require.NoError(t, err)

	err = repo.Adjust("acc-1", 1200)
	require.NoError(t, err)

	balance, err := repo.Get("acc-1")
	require.NoError(t, err)
	require.Equal(t, int64(1200), balance)
}

func TestBalanceRepositoryGetMissing(t *testing.T) {
	repo := NewBalanceRepository(cache.New())

	_, err := repo.Get("missing")
	require.ErrorIs(t, err, cache.ErrKeyNotFound)
}

func TestBalanceRepositoryAdjustDecrease(t *testing.T) {
	redis := cache.New()
	repo := NewBalanceRepository(redis)

	err := redis.Set("acc-2", "1000")
	require.NoError(t, err)

	err = repo.Adjust("acc-2", -250)
	require.NoError(t, err)

	balance, err := repo.Get("acc-2")
	require.NoError(t, err)
	require.Equal(t, int64(750), balance)
}

func TestBalanceRepositoryAdjustInsufficientBalance(t *testing.T) {
	redis := cache.New()
	repo := NewBalanceRepository(redis)

	err := redis.Set("acc-3", "100")
	require.NoError(t, err)

	err = repo.Adjust("acc-3", -150)
	require.ErrorIs(t, err, ErrInsufficientBalance)

	balance, getErr := repo.Get("acc-3")
	require.NoError(t, getErr)
	require.Equal(t, int64(100), balance)
}

func TestBalanceRepositoryAdjustMissingAccount(t *testing.T) {
	repo := NewBalanceRepository(cache.New())

	err := repo.Adjust("missing", 100)
	require.ErrorIs(t, err, cache.ErrKeyNotFound)
}
