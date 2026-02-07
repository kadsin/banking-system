package middlewares_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/server/middlewares"
	"github.com/stretchr/testify/require"
)

func TestErrorHandler_FiberError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler,
	})

	app.Get("/error-forbidden", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusForbidden, "denied")
	})

	req := httptest.NewRequest("GET", "/error-forbidden", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestErrorHandler_InternalDebugTrue(t *testing.T) {
	previous := config.Env.App.Debug
	config.Env.App.Debug = true
	defer func() {
		config.Env.App.Debug = previous
	}()

	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler,
	})

	app.Get("/error-internal-server", func(c *fiber.Ctx) error {
		return errors.New("boom")
	})

	req := httptest.NewRequest("GET", "/error-internal-server", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestErrorHandler_InternalDebugFalse(t *testing.T) {
	previous := config.Env.App.Debug
	config.Env.App.Debug = false
	defer func() {
		config.Env.App.Debug = previous
	}()

	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler,
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		return errors.New("boom")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
