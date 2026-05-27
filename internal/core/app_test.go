package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/queue"
	"github.com/stretchr/testify/require"
)

func TestPullAndProcessBulkInsertAndAdjustBalances(t *testing.T) {
	q := queue.New()
	txRepo := datalayer.NewMainTransactionRepository()
	ledgerRepo := datalayer.NewMainLedgerRepository()
	app := New(txRepo, ledgerRepo, q)

	topic := config.Env.Topics.Transactions

	_, err := q.Publish(topic, []byte(`{"id":"tx-1","from_account":"a","to_account":"b","amount":100}`))
	require.NoError(t, err)
	_, err = q.Publish(topic, []byte(`{"id":"tx-2","from_account":"b","to_account":"c","amount":40}`))
	require.NoError(t, err)

	n, err := app.PullAndProcess(10)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	txs, err := txRepo.List()
	require.NoError(t, err)
	require.Len(t, txs, 2)
	require.Equal(t, "tx-1", txs[0].ID)
	require.Equal(t, "tx-2", txs[1].ID)
	require.Equal(t, "COMPLETED", string(txs[0].Status))
	require.Equal(t, "COMPLETED", string(txs[1].Status))

	aBalance, err := ledgerRepo.Get("a")
	require.NoError(t, err)
	require.Equal(t, int64(-100), aBalance)

	bBalance, err := ledgerRepo.Get("b")
	require.NoError(t, err)
	require.Equal(t, int64(60), bBalance)

	cBalance, err := ledgerRepo.Get("c")
	require.NoError(t, err)
	require.Equal(t, int64(40), cBalance)

	remaining, err := q.Fetch(topic, 10)
	require.NoError(t, err)
	require.Len(t, remaining, 2)
}

func TestPullAndProcessInvalidPayloadDoesNotCommit(t *testing.T) {
	q := queue.New()
	txRepo := datalayer.NewMainTransactionRepository()
	ledgerRepo := datalayer.NewMainLedgerRepository()
	app := New(txRepo, ledgerRepo, q)

	topic := config.Env.Topics.Transactions

	_, err := q.Publish(topic, []byte(`not-json`))
	require.NoError(t, err)

	_, err = app.PullAndProcess(10)
	require.Error(t, err)

	messages, err := q.Fetch(topic, 10)
	require.NoError(t, err)
	require.Len(t, messages, 1)
}

func TestRunPollsAndProcesses(t *testing.T) {
	q := queue.New()
	txRepo := datalayer.NewMainTransactionRepository()
	ledgerRepo := datalayer.NewMainLedgerRepository()
	app := New(txRepo, ledgerRepo, q)

	topic := config.Env.Topics.Transactions

	_, err := q.Publish(topic, []byte(`{"id":"tx-10","from_account":"x","to_account":"y","amount":15}`))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	require.Eventually(t, func() bool {
		txs, err := txRepo.List()
		if err != nil {
			return false
		}

		return len(txs) == 1
	}, 500*time.Millisecond, 10*time.Millisecond)

	cancel()
	err = <-done
	require.ErrorIs(t, err, context.Canceled)
}

func TestPullAndProcessPublishesFailedEventOnAdjustmentError(t *testing.T) {
	q := queue.New()
	txRepo := datalayer.NewMainTransactionRepository()
	ledgerRepo := &failingLedgerRepository{}
	app := New(txRepo, ledgerRepo, q)

	txTopic := config.Env.Topics.Transactions
	failedTopic := config.Env.Topics.Failed

	_, err := q.Publish(txTopic, []byte(`{"id":"tx-11","from_account":"z","to_account":"w","amount":10}`))
	require.NoError(t, err)

	n, err := app.PullAndProcess(10)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	failed, err := q.Fetch(failedTopic, 10)
	require.NoError(t, err)
	require.Len(t, failed, 1)
}

func TestPullAndProcessIgnoresCompletedStatusEvents(t *testing.T) {
	q := queue.New()
	txRepo := datalayer.NewMainTransactionRepository()
	ledgerRepo := datalayer.NewMainLedgerRepository()
	app := New(txRepo, ledgerRepo, q)

	topic := config.Env.Topics.Transactions

	_, err := q.Publish(topic, []byte(`{"id":"tx-1","from_account":"a","to_account":"b","amount":100,"status":"PENDING"}`))
	require.NoError(t, err)
	_, err = q.Publish(topic, []byte(`{"id":"tx-1","from_account":"a","to_account":"b","amount":100,"status":"COMPLETED"}`))
	require.NoError(t, err)

	n, err := app.PullAndProcess(10)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	aBalance, err := ledgerRepo.Get("a")
	require.NoError(t, err)
	require.Equal(t, int64(-100), aBalance)

	bBalance, err := ledgerRepo.Get("b")
	require.NoError(t, err)
	require.Equal(t, int64(100), bBalance)
}

type failingLedgerRepository struct{}

func (f *failingLedgerRepository) Adjust(accountID string, delta int64) error {
	return errors.New("adjust failed")
}

func (f *failingLedgerRepository) Get(accountID string) (int64, error) {
	return 0, nil
}
