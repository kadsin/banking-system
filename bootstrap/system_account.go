package bootstrap

import (
	"log"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/domain"
)

func initSystemAccount(c container) {
	balance := config.Env.Treasury.InitialBalance

	if err := c.LedgerRepo.Set(domain.SystemAccountID, balance); err != nil {
		log.Fatalf("failed to init system account in ledger repo: %v", err)
	}

	if err := c.MainLedgerRepo.Adjust(domain.SystemAccountID, balance); err != nil {
		log.Fatalf("failed to init system account in main ledger repo: %v", err)
	}

	_, err := c.AccountRepo.Create(domain.Account{
		ID:       domain.SystemAccountID,
		Balance:  balance,
		Currency: "USD",
		Status:   domain.AccountStatusActive,
	})
	if err != nil {
		log.Fatalf("failed to init system account in account repo: %v", err)
	}
}
