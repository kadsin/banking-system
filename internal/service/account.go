package service

import (
	"fmt"

	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/domain"
)

func NewAccountService(accounts contracts.AccountRepository, balance contracts.BalanceService, transfer contracts.TransferService) *accountService {
	return &accountService{
		accounts: accounts,
		balance:  balance,
		transfer: transfer,
	}
}

type accountService struct {
	accounts contracts.AccountRepository
	balance  contracts.BalanceService
	transfer contracts.TransferService
}

func (s *accountService) Create(account domain.Account) (domain.Account, error) {
	if account.Balance < 0 {
		return domain.Account{}, fmt.Errorf("initial balance cannot be negative")
	}

	created, err := s.accounts.Create(account)
	if err != nil {
		return domain.Account{}, err
	}

	if created.Balance > 0 {
		_, err = s.transfer.Transfer(contracts.TransferInput{
			FromAccountID:  domain.SystemAccountID,
			ToAccountID:    created.ID,
			Amount:         created.Balance,
			IdempotencyKey: "init-balance:" + created.ID,
		})
		if err != nil {
			return domain.Account{}, err
		}
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
