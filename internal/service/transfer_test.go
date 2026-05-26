package service

import (
	"encoding/json"
	"testing"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestTransferServiceCreatesPendingOutboxEvent(t *testing.T) {
	svc, deps := newTransferServiceTestDeps(t)

	fromID := "11111111-1111-1111-1111-111111111111"
	toID := "22222222-2222-2222-2222-222222222222"

	createAccount(t, deps.accounts, fromID, 5000, domain.AccountStatusActive)
	require.NoError(t, deps.ledger.Set(fromID, 5000))

	createAccount(t, deps.accounts, toID, 1000, domain.AccountStatusActive)
	require.NoError(t, deps.ledger.Set(toID, 1000))

	tx, err := svc.Transfer(contracts.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         1500,
		IdempotencyKey: "idem-1",
	})
	require.NoError(t, err)
	require.Equal(t, domain.TransactionStatusPending, tx.Status)
	require.Equal(t, fromID, tx.FromAccountID)
	require.Equal(t, toID, tx.ToAccountID)
	require.Equal(t, int64(1500), tx.Amount)
	require.Equal(t, "idem-1", tx.IdempotencyKey)
	require.NotEmpty(t, tx.ID)

	from, err := deps.accounts.GetByID(fromID)
	require.NoError(t, err)
	require.Equal(t, int64(3500), from.Balance)

	to, err := deps.accounts.GetByID(toID)
	require.NoError(t, err)
	require.Equal(t, int64(2500), to.Balance)

	events, err := deps.outbox.ListPending(10)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, config.Env.Topics.Transactions, events[0].Topic)

	var outboxTx domain.Transaction
	require.NoError(t, json.Unmarshal(events[0].Payload, &outboxTx))
	require.Equal(t, tx, outboxTx)

	storedTxID, exists, err := deps.idempotencies.Get("idem-1")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, tx.ID, storedTxID)
}

func TestTransferServiceReturnsExistingTransactionForIdempotencyKey(t *testing.T) {
	svc, deps := newTransferServiceTestDeps(t)

	fromID := "11111111-1111-1111-1111-111111111111"
	toID := "22222222-2222-2222-2222-222222222222"

	createAccount(t, deps.accounts, fromID, 0, domain.AccountStatusActive)
	require.NoError(t, deps.ledger.Set(fromID, 3000))

	createAccount(t, deps.accounts, toID, 0, domain.AccountStatusActive)
	require.NoError(t, deps.ledger.Set(toID, 3000))

	first, err := svc.Transfer(contracts.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         500,
		IdempotencyKey: "idem-replay",
	})
	require.NoError(t, err)

	_, err = deps.transactions.Create(first)
	require.NoError(t, err)

	second, err := svc.Transfer(contracts.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         500,
		IdempotencyKey: "idem-replay",
	})
	require.NoError(t, err)
	require.Equal(t, first, second)

	events, err := deps.outbox.ListPending(10)
	require.NoError(t, err)
	require.Len(t, events, 1)
}

func TestTransferServiceRejectsMissingAccount(t *testing.T) {
	svc, deps := newTransferServiceTestDeps(t)

	fromID := "11111111-1111-1111-1111-111111111111"
	toID := "22222222-2222-2222-2222-222222222222"

	createAccount(t, deps.accounts, fromID, 1000, domain.AccountStatusActive)

	_, err := svc.Transfer(contracts.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         100,
		IdempotencyKey: "idem-missing-to",
	})
	require.ErrorIs(t, err, datalayer.ErrAccountNotFound)

	events, err := deps.outbox.ListPending(10)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestTransferServiceRejectsBlockedAccount(t *testing.T) {
	svc, deps := newTransferServiceTestDeps(t)

	fromID := "11111111-1111-1111-1111-111111111111"
	toID := "22222222-2222-2222-2222-222222222222"

	createAccount(t, deps.accounts, fromID, 1000, domain.AccountStatusBlocked)
	createAccount(t, deps.accounts, toID, 1000, domain.AccountStatusActive)

	_, err := svc.Transfer(contracts.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         100,
		IdempotencyKey: "idem-blocked-from",
	})
	require.ErrorIs(t, err, ErrAccountBlocked)

	createAccount(t, deps.accounts, fromID, 1000, domain.AccountStatusActive)
	createAccount(t, deps.accounts, toID, 1000, domain.AccountStatusBlocked)

	_, err = svc.Transfer(contracts.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         100,
		IdempotencyKey: "idem-blocked-to",
	})
	require.ErrorIs(t, err, ErrAccountBlocked)

	events, err := deps.outbox.ListPending(10)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestTransferServiceRejectsInsufficientFunds(t *testing.T) {
	svc, deps := newTransferServiceTestDeps(t)

	fromID := "11111111-1111-1111-1111-111111111111"
	toID := "22222222-2222-2222-2222-222222222222"

	createAccount(t, deps.accounts, fromID, 100, domain.AccountStatusActive)
	createAccount(t, deps.accounts, toID, 0, domain.AccountStatusActive)
	require.NoError(t, deps.ledger.Set(fromID, 100))

	_, err := svc.Transfer(contracts.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         500,
		IdempotencyKey: "idem-insufficient",
	})
	require.ErrorIs(t, err, ErrInsufficientFunds)

	events, err := deps.outbox.ListPending(10)
	require.NoError(t, err)
	require.Empty(t, events)

	_, exists, err := deps.idempotencies.Get("idem-insufficient")
	require.NoError(t, err)
	require.False(t, exists)
}

type transferServiceTestDeps struct {
	accounts      *datalayer.AccountRepository
	transactions  *datalayer.OlapRepository
	ledger        *datalayer.LedgerRepository
	outbox        *datalayer.OutboxRepository
	idempotencies *datalayer.TxIdempotencyRepository
}

func newTransferServiceTestDeps(t *testing.T) (*transferService, transferServiceTestDeps) {
	t.Helper()

	transactions := datalayer.NewOlapRepository(nil)
	ledger := datalayer.NewLedgerRepository(cache.New())
	balance := NewBalanceService(ledger)
	accounts := datalayer.NewAccountRepository(balance)
	outbox := datalayer.NewOutboxRepository()
	idempotencies := datalayer.NewTxIdempotencyRepository(cache.New())

	svc := NewTransferService(
		accounts,
		transactions,
		balance,
		outbox,
		idempotencies,
	)

	return svc, transferServiceTestDeps{
		accounts:      accounts,
		transactions:  transactions,
		ledger:        ledger,
		outbox:        outbox,
		idempotencies: idempotencies,
	}
}

func createAccount(t *testing.T, repo *datalayer.AccountRepository, id string, balance int64, status domain.AccountStatus) {
	t.Helper()

	_, err := repo.Create(domain.Account{
		ID:       id,
		Balance:  balance,
		Currency: "USD",
		Status:   status,
	})
	require.NoError(t, err)
}
