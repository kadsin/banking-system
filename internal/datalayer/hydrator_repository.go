package datalayer

import (
	"encoding/json"
	"errors"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/contracts"
)

var ErrInvalidSnapshotValue = errors.New("invalid snapshot value")

type HydratorRepository struct {
	cache *cache.Cache
}

func NewHydratorRepository(snapshotStore *cache.Cache) *HydratorRepository {
	return &HydratorRepository{cache: snapshotStore}
}

func (r *HydratorRepository) SetSnapshot(accountID string, snapshot contracts.HydratorSnapshot) error {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	return r.cache.Set(accountID, string(payload))
}

func (r *HydratorRepository) GetSnapshot(accountID string) (contracts.HydratorSnapshot, error) {
	value, err := r.cache.Get(accountID)
	if err != nil {
		return contracts.HydratorSnapshot{}, err
	}

	var snapshot contracts.HydratorSnapshot
	err = json.Unmarshal([]byte(value), &snapshot)
	if err != nil {
		return contracts.HydratorSnapshot{}, ErrInvalidSnapshotValue
	}

	return snapshot, nil
}
