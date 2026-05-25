package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountBlocked    = errors.New("account is blocked")
)

func NewTransferService(accounts contracts.AccountRepository, transactions contracts.TransactionRepository) *transferService {
	return &transferService{
		accounts:     accounts,
		transactions: transactions,
	}
}

type transferService struct {
	accounts     contracts.AccountRepository
	transactions contracts.TransactionRepository
}

func (s *transferService) Transfer(input contracts.TransferInput) (domain.Transaction, error) {
	if existing, ok, err := s.transactions.GetByIdempotencyKey(input.IdempotencyKey); err != nil {
		return domain.Transaction{}, err
	} else if ok {
		return existing, nil
	}

	from, err := s.accounts.GetByID(input.FromAccountID)
	if err != nil {
		return domain.Transaction{}, err
	}

	to, err := s.accounts.GetByID(input.ToAccountID)
	if err != nil {
		return domain.Transaction{}, err
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

	transaction, err = s.transactions.Create(transaction)
	if err != nil {
		return domain.Transaction{}, err
	}

	if from.Status == domain.AccountStatusBlocked || to.Status == domain.AccountStatusBlocked {
		_ = s.transactions.UpdateStatus(transaction.ID, domain.TransactionStatusFailed)
		return domain.Transaction{}, ErrAccountBlocked
	}

	if from.Balance < input.Amount {
		_ = s.transactions.UpdateStatus(transaction.ID, domain.TransactionStatusFailed)
		return domain.Transaction{}, ErrInsufficientFunds
	}

	from.Balance -= input.Amount
	if err := s.accounts.Update(from); err != nil {
		return domain.Transaction{}, err
	}

	to.Balance += input.Amount
	if err := s.accounts.Update(to); err != nil {
		return domain.Transaction{}, err
	}

	if err := s.transactions.UpdateStatus(transaction.ID, domain.TransactionStatusCompleted); err != nil {
		return domain.Transaction{}, err
	}

	transaction.Status = domain.TransactionStatusCompleted

	return transaction, nil
}
