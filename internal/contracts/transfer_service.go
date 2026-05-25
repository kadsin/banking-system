package contracts

import (
	"github.com/kadsin/banking-system/internal/domain"
)

type TransferService interface {
	Transfer(input TransferInput) (domain.Transaction, error)
}

type TransferInput struct {
	FromAccountID  string
	ToAccountID    string
	Amount         int64
	IdempotencyKey string
}
