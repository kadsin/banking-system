package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/swagger"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/server/handlers"
	"github.com/kadsin/banking-system/internal/server/middlewares"
)

type Dependencies struct {
	Accounts   contracts.AccountRepository
	Txs        contracts.TransactionRepository
	Transferer contracts.TransferService
	Hydrator   contracts.HydratorService
}

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	setupSwagger(app.Group("/docs"))

	api := app.Group("/api")

	authHandler := handlers.NewAuthHandler()
	api.Post("/login", authHandler.Login)

	auth := api.Group("", middlewares.RequireAuth)

	accountHandler := handlers.NewAccountHandler(deps.Accounts)
	auth.Post("/accounts", accountHandler.Create)
	auth.Get("/accounts/:id", accountHandler.GetByID)

	transactionHandler := handlers.NewTransactionHandler(deps.Txs, deps.Transferer)
	auth.Post("/transactions/transfer", transactionHandler.Transfer)
	auth.Get("/transactions/:id", transactionHandler.GetByID)
	auth.Get("/transactions/:account_id/history", transactionHandler.History)

	// We assume that the internal routes are not visible to end users
	internal := app.Group("/internal")
	hydratorHandler := handlers.NewHydratorHandler(deps.Hydrator)
	internal.Post("/hydrator/:account_id/repopulate", hydratorHandler.Repopulate)
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
