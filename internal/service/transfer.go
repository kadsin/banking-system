package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountBlocked    = errors.New("account is blocked")
)

func NewTransferService(accounts contracts.AccountRepository, transactions contracts.OlapRepository, balance contracts.BalanceService, outbox contracts.OutboxRepository, txIdempotency contracts.TxIdempotencyRepository) *transferService {
	return &transferService{
		accounts:      accounts,
		transactions:  transactions,
		balance:       balance,
		outbox:        outbox,
		idempotencies: txIdempotency,
	}
}

type transferService struct {
	accounts      contracts.AccountRepository
	transactions  contracts.OlapRepository
	balance       contracts.BalanceService
	outbox        contracts.OutboxRepository
	idempotencies contracts.TxIdempotencyRepository
}

func (s *transferService) Transfer(input contracts.TransferInput) (domain.Transaction, error) {
	if existingTxID, ok, err := s.idempotencies.Get(input.IdempotencyKey); err != nil {
		return domain.Transaction{}, err
	} else if ok {
		return s.transactions.GetByID(existingTxID)
	}

	from, err := s.accounts.GetByID(input.FromAccountID)
	if err != nil {
		return domain.Transaction{}, err
	}

	to, err := s.accounts.GetByID(input.ToAccountID)
	if err != nil {
		return domain.Transaction{}, err
	}

	if from.Status == domain.AccountStatusBlocked || to.Status == domain.AccountStatusBlocked {
		return domain.Transaction{}, ErrAccountBlocked
	}

	transaction := domain.Transaction{
		ID:             uuid.NewString(),
		FromAccountID:  input.FromAccountID,
		ToAccountID:    input.ToAccountID,
		Amount:         input.Amount,
		Status:         domain.TransactionStatusPending,
		IdempotencyKey: input.IdempotencyKey,
		Timestamp:      time.Now().UTC(),
	}

	if fromBalance, err := s.balance.Get(input.FromAccountID); err != nil {
		return domain.Transaction{}, err
	} else if fromBalance < input.Amount {
		return domain.Transaction{}, ErrInsufficientFunds
	}

	payload, err := json.Marshal(transaction)
	if err != nil {
		return domain.Transaction{}, err
	}

	if _, err := s.outbox.Create(config.Env.Topics.Transactions, payload); err != nil {
		return domain.Transaction{}, err
	}

	if err := s.balance.Adjust(input.FromAccountID, -input.Amount); err != nil {
		return domain.Transaction{}, err
	}
	if err := s.balance.Adjust(input.ToAccountID, input.Amount); err != nil {
		return domain.Transaction{}, err
	}

	if err := s.idempotencies.Set(input.IdempotencyKey, transaction.ID); err != nil {
		return domain.Transaction{}, err
	}

	return transaction, nil
}
