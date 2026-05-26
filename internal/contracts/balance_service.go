package contracts

import "github.com/kadsin/banking-system/internal/domain"

type BalanceService interface {
	Get(accountID string) (int64, error)
	Adjust(accountID string, delta int64) error
	Refund(transaction domain.Transaction) error
}
