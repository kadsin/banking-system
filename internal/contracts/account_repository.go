package contracts

import "github.com/kadsin/banking-system/internal/domain"

type AccountRepository interface {
	Create(account domain.Account) (domain.Account, error)
	GetByID(id string) (domain.Account, error)
	Update(account domain.Account) error
}
