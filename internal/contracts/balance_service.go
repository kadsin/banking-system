package contracts

import "github.com/kadsin/banking-system/internal/domain"

type BalanceService interface {
	Refund(transaction domain.Transaction) error
}
