package datalayer

import "sync"

func NewMainLedgerRepository() *MainLedgerRepository {
	return &MainLedgerRepository{
		ledgers: map[string]int64{},
	}
}

type MainLedgerRepository struct {
	mu      sync.RWMutex
	ledgers map[string]int64
}

func (r *MainLedgerRepository) Adjust(accountID string, delta int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ledgers[accountID] += delta
	return nil
}

func (r *MainLedgerRepository) Get(accountID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.ledgers[accountID], nil
}
