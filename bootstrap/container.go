package bootstrap

import (
	"context"
	"log"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/cdc"
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

	HydratorRepo  contracts.HydratorRepository
	HydratorCache *cache.Cache
	Hydrator      contracts.HydratorService

	TxIdempotencyCache *cache.Cache
	TransferService    contracts.TransferService
	TransactionService contracts.TransactionService
	TxIdempotencyRepo  contracts.TxIdempotencyRepository
	OlapRepo           contracts.OlapRepository
	OutboxRepo         contracts.OutboxRepository

	CoreWorker     *core.App
	CDCWorker      *cdc.App
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
	log.Println("Starting core worker")
	go func() {
		if err := coreWorker.Run(ctx); err != nil {
			log.Fatalf("core worker stopped: %v", err)
		}
	}()

	ledgerCache := cache.New()
	ledgerRepo := datalayer.NewLedgerRepository(ledgerCache)

	hydratorCache := cache.New()
	hydratorRepo := datalayer.NewHydratorRepository(hydratorCache)
	hydratorService := service.NewHydratorService(ledgerRepo, hydratorRepo, q)

	balanceService := service.NewBalanceService(ledgerRepo, hydratorService)

	sagaWorker := saga.New(balanceService, q)
	log.Println("Starting saga worker")
	go func() {
		if err := sagaWorker.Run(ctx); err != nil {
			log.Fatalf("saga worker stopped: %v", err)
		}
	}()

	accountRepo := datalayer.NewAccountRepository()
	olapRepo := datalayer.NewOlapRepository(q)

	outboxRepo := datalayer.NewOutboxRepository()

	cdcWorker := cdc.New(outboxRepo, q)
	log.Println("Starting cdc worker")
	go func() {
		if err := cdcWorker.Run(ctx); err != nil {
			log.Fatalf("cdc worker stopped: %v", err)
		}
	}()

	txIdempotencyCache := cache.New()
	txIdempotencyRepo := datalayer.NewTxIdempotencyRepository(txIdempotencyCache)

	transferService := service.NewTransferService(accountRepo, olapRepo, balanceService, outboxRepo, txIdempotencyRepo)
	transactionService := service.NewTransactionService(olapRepo)

	accountService := service.NewAccountService(accountRepo, balanceService, transferService)

	c := container{
		Queue: q,

		LedgerCache:    ledgerCache,
		LedgerRepo:     ledgerRepo,
		BalanceService: balanceService,

		HydratorRepo:  hydratorRepo,
		HydratorCache: hydratorCache,
		Hydrator:      hydratorService,

		TxIdempotencyCache: txIdempotencyCache,
		TransferService:    transferService,
		TransactionService: transactionService,
		TxIdempotencyRepo:  txIdempotencyRepo,
		OlapRepo:           olapRepo,
		OutboxRepo:         outboxRepo,

		CoreWorker:     coreWorker,
		CDCWorker:      cdcWorker,
		MainTxRepo:     mainTxRepo,
		MainLedgerRepo: mainLedgerRepo,

		SagaWorker: sagaWorker,

		AccountRepo:    accountRepo,
		AccountService: accountService,
	}

	log.Println("Container initialized successfully")

	initSystemAccount(c)
	log.Println("System account initialized successfully")

	return c
}
