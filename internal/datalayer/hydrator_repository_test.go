package datalayer

import (
	"testing"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/stretchr/testify/require"
)

func TestHydratorRepositorySetAndGetSnapshot(t *testing.T) {
	repo := NewHydratorRepository(cache.New())

	err := repo.SetSnapshot("acc-1", contracts.HydratorSnapshot{
		Balance:     2200,
		QueueOffset: 45,
	})
	require.NoError(t, err)

	snapshot, err := repo.GetSnapshot("acc-1")
	require.NoError(t, err)
	require.Equal(t, int64(2200), snapshot.Balance)
	require.Equal(t, int64(45), snapshot.QueueOffset)
}

func TestHydratorRepositoryGetSnapshotMissing(t *testing.T) {
	repo := NewHydratorRepository(cache.New())

	_, err := repo.GetSnapshot("missing")
	require.ErrorIs(t, err, cache.ErrKeyNotFound)
}

func TestHydratorRepositoryGetSnapshotInvalidValue(t *testing.T) {
	redis := cache.New()
	repo := NewHydratorRepository(redis)

	err := redis.Set("acc-2", "not-a-number")
	require.NoError(t, err)

	_, err = repo.GetSnapshot("acc-2")
	require.ErrorIs(t, err, ErrInvalidSnapshotValue)
}
