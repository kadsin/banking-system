package service

import (
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
)

func NewAccountService(accounts contracts.AccountRepository, balance contracts.BalanceService) *accountService {
	return &accountService{
		accounts: accounts,
		balance:  balance,
	}
}

type accountService struct {
	accounts contracts.AccountRepository
	balance  contracts.BalanceService
}

func (s *accountService) Create(account domain.Account) (domain.Account, error) {
	created, err := s.accounts.Create(account)
	if err != nil {
		return domain.Account{}, err
	}

	if err := s.balance.Adjust(created.ID, created.Balance); err != nil {
		return domain.Account{}, err
	}

	return created, nil
}

func (s *accountService) GetByID(id string) (domain.Account, error) {
	account, err := s.accounts.GetByID(id)
	if err != nil {
		return domain.Account{}, err
	}

	balance, err := s.balance.Get(id)
	if err != nil {
		return domain.Account{}, err
	}

	account.Balance = balance
	return account, nil
}
