package contracts

import "github.com/kadsin/banking-system/internal/domain"

type AccountService interface {
	Create(account domain.Account) (domain.Account, error)
	GetByID(id string) (domain.Account, error)
}
