package bootstrap

import (
	"context"
	"log"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/core"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/queue"
	"github.com/kadsin/banking-system/internal/saga"
	"github.com/kadsin/banking-system/internal/service"
)

type container struct {
	Queue *queue.Queue

	LedgerCache    *cache.Cache
	LedgerRepo     contracts.LedgerRepository
	BalanceService contracts.BalanceService

	TxIdempotencyCache *cache.Cache
	TransferService    contracts.TransferService
	TransactionService contracts.TransactionService
	TxIdempotencyRepo  contracts.TxIdempotencyRepository
	OlapRepo           contracts.OlapRepository
	OutboxRepo         contracts.OutboxRepository

	CoreWorker     *core.App
	MainTxRepo     contracts.MainTransactionRepository
	MainLedgerRepo contracts.MainLedgerRepository

	SagaWorker *saga.App

	AccountRepo    contracts.AccountRepository
	AccountService contracts.AccountService
}

func InitContainer(ctx context.Context) container {
	q := queue.New()

	mainTxRepo := datalayer.NewMainTransactionRepository()
	mainLedgerRepo := datalayer.NewMainLedgerRepository()

	coreWorker := core.New(mainTxRepo, mainLedgerRepo, q)
	go func() {
		if err := coreWorker.Run(ctx); err != nil {
			log.Fatalf("core worker stopped: %v", err)
		}
	}()

	ledgerCache := cache.New()
	ledgerRepo := datalayer.NewLedgerRepository(ledgerCache)
	balanceService := service.NewBalanceService(ledgerRepo)

	sagaWorker := saga.New(balanceService, q)
	go func() {
		if err := sagaWorker.Run(ctx); err != nil {
			log.Fatalf("saga worker stopped: %v", err)
		}
	}()

	accountRepo := datalayer.NewAccountRepository()
	accountService := service.NewAccountService(accountRepo, balanceService)
	olapRepo := datalayer.NewOlapRepository(q)

	outboxRepo := datalayer.NewOutboxRepository()

	txIdempotencyCache := cache.New()
	txIdempotencyRepo := datalayer.NewTxIdempotencyRepository(txIdempotencyCache)

	transferService := service.NewTransferService(accountRepo, olapRepo, balanceService, outboxRepo, txIdempotencyRepo)
	transactionService := service.NewTransactionService(olapRepo)

	return container{
		Queue: q,

		LedgerCache:    ledgerCache,
		LedgerRepo:     ledgerRepo,
		BalanceService: balanceService,

		TxIdempotencyCache: txIdempotencyCache,
		TransferService:    transferService,
		TransactionService: transactionService,
		TxIdempotencyRepo:  txIdempotencyRepo,
		OlapRepo:           olapRepo,
		OutboxRepo:         outboxRepo,

		CoreWorker:     coreWorker,
		MainTxRepo:     mainTxRepo,
		MainLedgerRepo: mainLedgerRepo,

		SagaWorker: sagaWorker,

		AccountRepo:    accountRepo,
		AccountService: accountService,
	}
}
