package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/swagger"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/server/handlers"
)

type Dependencies struct {
	Accounts   contracts.AccountRepository
	Txs        contracts.TransactionRepository
	Transferer contracts.TransferService
}

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	setupSwagger(app.Group("/docs"))

	api := app.Group("/api")

	accountHandler := handlers.NewAccountHandler(deps.Accounts)
	api.Post("/accounts", accountHandler.Create)
	api.Get("/accounts/:id", accountHandler.GetByID)

	transactionHandler := handlers.NewTransactionHandler(deps.Txs, deps.Transferer)
	api.Post("/transactions/transfer", transactionHandler.Transfer)
	api.Get("/transactions/:id", transactionHandler.GetByID)
}

func setupSwagger(router fiber.Router) {
	router.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			config.Env.Doc.Auth.Username: config.Env.Doc.Auth.Password,
		},
	}))

	router.Static("/swagger.yml", "./docs/swagger.yml")

	router.Get("/*", swagger.New(swagger.Config{
		Title: config.Env.App.Name + " - API Doc",
		URL:   "/docs/swagger.yml",
	}))
}
