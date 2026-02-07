package requests_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/kadsin/banking-system/internal/server/requests"
	"github.com/stretchr/testify/require"
)

type preparePayload struct {
	Name string `json:"name" validate:"required"`
}

func TestPrepare_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/payload", func(c *fiber.Ctx) error {
		body, err := requests.Prepare[preparePayload](c)

		require.NoError(t, err)
		require.Equal(t, "john", body.Name)

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/payload", strings.NewReader(`{"name":"john"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestPrepare_InvalidJSON(t *testing.T) {
	app := fiber.New()
	app.Post("/payload", func(c *fiber.Ctx) error {
		_, err := requests.Prepare[preparePayload](c)
		require.Error(t, err)

		return c.SendStatus(fiber.StatusBadRequest)
	})

	req := httptest.NewRequest("POST", "/payload", strings.NewReader(`bad json`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestPrepare_ValidationError(t *testing.T) {
	app := fiber.New()
	app.Post("/payload", func(c *fiber.Ctx) error {
		_, err := requests.Prepare[preparePayload](c)
		require.Error(t, err)

		return c.SendStatus(fiber.StatusUnprocessableEntity)
	})

	req := httptest.NewRequest("POST", "/payload", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}
