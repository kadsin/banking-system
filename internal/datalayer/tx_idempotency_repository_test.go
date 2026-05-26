package datalayer

import (
	"testing"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestTxIdempotencyRepositoryGetAndSet(t *testing.T) {
	redis := cache.New()
	repo := NewTxIdempotencyRepository(redis)

	txID, exists, err := repo.Get("k1")
	require.NoError(t, err)
	require.False(t, exists)
	require.Empty(t, txID)

	err = repo.Set("k1", "tx-1")
	require.NoError(t, err)

	txID, exists, err = repo.Get("k1")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "tx-1", txID)
}
