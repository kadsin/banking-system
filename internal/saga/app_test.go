package saga

import (
	"context"
	"testing"
	"time"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/queue"
	"github.com/kadsin/banking-system/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRunSagaAndRefunds(t *testing.T) {
	q := queue.New()
	ledger := datalayer.NewMainLedgerRepository()
	balanceService := service.NewBalanceService(ledger)
	app := New(balanceService, q)

	_, err := q.Publish(
		config.Env.Topics.Failed,
		[]byte(`{"id":"tx-2","from_account":"x","to_account":"y","amount":50}`),
	)

	require.NoError(t, err)
	require.NoError(t, ledger.Adjust("x", 100))
	require.NoError(t, ledger.Adjust("y", 100))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	require.Eventually(t, func() bool {
		x, err := ledger.Get("x")
		if err != nil {
			return false
		}

		y, err := ledger.Get("y")
		if err != nil {
			return false
		}

		return x == 150 && y == 50
	}, 500*time.Millisecond, 10*time.Millisecond)

	cancel()
	err = <-done
	require.ErrorIs(t, err, context.Canceled)
}
