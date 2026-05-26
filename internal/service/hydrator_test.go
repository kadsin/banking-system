package service

import (
	"encoding/json"
	"testing"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/kadsin/banking-system/internal/queue"
	"github.com/stretchr/testify/require"
)

func TestHydratorServiceRepopulateFromSnapshotAndQueueWithoutCommit(t *testing.T) {
	q := queue.New()
	ledgerRepo := datalayer.NewLedgerRepository(cache.New())
	hydratorRepo := datalayer.NewHydratorRepository(cache.New())
	svc := NewHydratorService(ledgerRepo, hydratorRepo, q)

	accountID := "11111111-1111-1111-1111-111111111111"
	otherID := "22222222-2222-2222-2222-222222222222"

	err := hydratorRepo.SetSnapshot(accountID, contracts.HydratorSnapshot{
		Balance:     1000,
		QueueOffset: 0,
	})
	require.NoError(t, err)

	topic := config.Env.Topics.Transactions

	_, err = publishTx(q, topic, domain.Transaction{FromAccountID: accountID, ToAccountID: otherID, Amount: 50}) // offset 0
	require.NoError(t, err)
	_, err = publishTx(q, topic, domain.Transaction{FromAccountID: otherID, ToAccountID: accountID, Amount: 200}) // offset 1
	require.NoError(t, err)
	_, err = publishTx(q, topic, domain.Transaction{FromAccountID: accountID, ToAccountID: otherID, Amount: 30}) // offset 2
	require.NoError(t, err)

	balance, err := svc.Repopulate(accountID)
	require.NoError(t, err)
	require.Equal(t, int64(1170), balance)

	stored, err := ledgerRepo.Get(accountID)
	require.NoError(t, err)
	require.Equal(t, int64(1170), stored)

	snapshot, err := hydratorRepo.GetSnapshot(accountID)
	require.NoError(t, err)
	require.Equal(t, int64(1170), snapshot.Balance)
	require.Equal(t, int64(2), snapshot.QueueOffset)

	messages, err := q.Fetch(topic, 10)
	require.NoError(t, err)
	require.Len(t, messages, 3)
	require.Equal(t, int64(0), messages[0].Offset)
}

func publishTx(q *queue.Queue, topic string, tx domain.Transaction) (queue.Message, error) {
	payload, err := json.Marshal(tx)
	if err != nil {
		return queue.Message{}, err
	}

	return q.Publish(topic, payload)
}
