package service

import (
	"errors"
	"testing"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestBalanceServiceGetCacheMissUsesHydrator(t *testing.T) {
	ledger := datalayer.NewLedgerRepository(cache.New())
	hydrator := &hydratorStub{
		repopulateFn: func(accountID string) (int64, error) {
			if accountID != "acc-1" {
				return 0, errors.New("unexpected account id")
			}

			if err := ledger.Set(accountID, 250); err != nil {
				return 0, err
			}

			return 250, nil
		},
	}

	svc := NewBalanceService(ledger, hydrator)
	balance, err := svc.Get("acc-1")
	require.NoError(t, err)
	require.Equal(t, int64(250), balance)
	require.Equal(t, 1, hydrator.calls)
}

func TestBalanceServiceGetCacheMissHydratorError(t *testing.T) {
	ledger := datalayer.NewLedgerRepository(cache.New())
	expectedErr := errors.New("hydrate failed")
	hydrator := &hydratorStub{
		repopulateFn: func(accountID string) (int64, error) {
			return 0, expectedErr
		},
	}

	svc := NewBalanceService(ledger, hydrator)
	_, err := svc.Get("acc-1")
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, 1, hydrator.calls)
}

func TestBalanceServiceAdjustCacheMissUsesHydrator(t *testing.T) {
	ledger := datalayer.NewLedgerRepository(cache.New())
	hydrator := &hydratorStub{
		repopulateFn: func(accountID string) (int64, error) {
			if err := ledger.Set(accountID, 100); err != nil {
				return 0, err
			}

			return 100, nil
		},
	}

	svc := NewBalanceService(ledger, hydrator)
	err := svc.Adjust("acc-2", -30)
	require.NoError(t, err)
	require.Equal(t, 1, hydrator.calls)

	balance, err := ledger.Get("acc-2")
	require.NoError(t, err)
	require.Equal(t, int64(70), balance)
}

func TestBalanceServiceRefundCacheMissUsesHydrator(t *testing.T) {
	ledger := datalayer.NewLedgerRepository(cache.New())
	hydrator := &hydratorStub{
		repopulateFn: func(accountID string) (int64, error) {
			if accountID == "from-acc" {
				if err := ledger.Set(accountID, 10); err != nil {
					return 0, err
				}
				return 10, nil
			}

			if accountID == "to-acc" {
				if err := ledger.Set(accountID, 40); err != nil {
					return 0, err
				}
				return 40, nil
			}

			return 0, errors.New("unexpected account")
		},
	}

	svc := NewBalanceService(ledger, hydrator)
	err := svc.Refund(domain.Transaction{
		FromAccountID: "from-acc",
		ToAccountID:   "to-acc",
		Amount:        15,
	})
	require.NoError(t, err)
	require.Equal(t, 2, hydrator.calls)

	fromBalance, err := ledger.Get("from-acc")
	require.NoError(t, err)
	require.Equal(t, int64(25), fromBalance)

	toBalance, err := ledger.Get("to-acc")
	require.NoError(t, err)
	require.Equal(t, int64(25), toBalance)
}

type hydratorStub struct {
	calls        int
	repopulateFn func(accountID string) (int64, error)
}

func (h *hydratorStub) Repopulate(accountID string) (int64, error) {
	h.calls++
	return h.repopulateFn(accountID)
}
