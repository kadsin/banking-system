package cdc

import (
	"context"
	"testing"
	"time"

	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/queue"
	"github.com/stretchr/testify/require"
)

func TestPullAndPushPublishesAndMarksProcessed(t *testing.T) {
	outbox := datalayer.NewOutboxRepository()
	q := queue.New()
	app := New(outbox, q)

	first, err := outbox.Create("transactions", []byte(`{"id":"tx-1"}`))
	require.NoError(t, err)

	second, err := outbox.Create("transactions", []byte(`{"id":"tx-2"}`))
	require.NoError(t, err)

	n, err := app.PullAndPush(10)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	msgs, err := q.Fetch("transactions", 10)
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte(`{"id":"tx-1"}`), msgs[0].Value)
	require.Equal(t, []byte(`{"id":"tx-2"}`), msgs[1].Value)

	err = q.Commit("transactions", msgs[1].Offset)
	require.NoError(t, err)

	rest, err := q.Fetch("transactions", 10)
	require.NoError(t, err)
	require.Len(t, rest, 0)

	firstStored, err := outbox.GetByID(first.ID)
	require.NoError(t, err)
	require.NotNil(t, firstStored.ProcessedAt)

	secondStored, err := outbox.GetByID(second.ID)
	require.NoError(t, err)
	require.NotNil(t, secondStored.ProcessedAt)
}

func TestRunWorkerPollsAndPublishes(t *testing.T) {
	outbox := datalayer.NewOutboxRepository()
	q := queue.New()
	app := New(outbox, q)

	_, err := outbox.Create("transactions", []byte(`{"id":"tx-10"}`))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	require.Eventually(t, func() bool {
		msgs, err := q.Fetch("transactions", 10)
		if err != nil {
			return false
		}

		return len(msgs) == 1
	}, 500*time.Millisecond, 10*time.Millisecond)

	cancel()
	err = <-done
	require.ErrorIs(t, err, context.Canceled)
}
