package middlewares_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/kadsin/banking-system/internal/server/middlewares"
	"github.com/stretchr/testify/require"
)

func TestResponseWrapper_WrapsJSONPayload(t *testing.T) {
	app := fiber.New()
	app.Use(middlewares.ResponseWrapper)
	app.Get("/json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": "world"})
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	_, ok := body["data"]
	require.True(t, ok)
}

func TestResponseWrapper_WrapsError(t *testing.T) {
	app := fiber.New()
	app.Use(middlewares.ResponseWrapper)
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(http.StatusBadRequest, "bad request")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	_, ok := body["errors"]
	require.True(t, ok)
}
