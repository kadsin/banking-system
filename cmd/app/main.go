package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kadsin/banking-system/bootstrap"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/server"
)

func main() {
	container := bootstrap.InitContainer(context.TODO())

	app := bootstrap.SetupFiberApp(&server.Dependencies{
		Accounts:     container.AccountService,
		Transactions: container.TransactionService,
		Transferer:   container.TransferService,
	})
	addr := fmt.Sprintf(":%s", config.Env.App.Port)

	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server listen failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("trying to graceful shutdown")

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
