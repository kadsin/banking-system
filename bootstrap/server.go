package bootstrap

import (
	"time"

	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/queue"
	"github.com/kadsin/banking-system/internal/server"
	"github.com/kadsin/banking-system/internal/server/middlewares"
	"github.com/kadsin/banking-system/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

func SetupFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler,
	})

	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(recover.New())

	app.Use("/api",
		cors.New(),
		timeout.NewWithContext(func(c *fiber.Ctx) error { return c.Next() }, 8*time.Second),
		middlewares.ResponseWrapper,
	)

	server.SetupRoutes(app, resolveServerDependencies())

	return app
}

func resolveServerDependencies() *server.Dependencies {
	q := queue.New()

	accounts := datalayer.NewAccountRepository()
	txs := datalayer.NewOlapRepository(q)

	return &server.Dependencies{
		Accounts: accounts,
		Txs:      txs,
		Transferer: service.NewTransferService(
			accounts,
			txs,
			datalayer.NewTxIdempotencyRepository(cache.New()),
		),
	}
}
