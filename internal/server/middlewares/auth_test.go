package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/auth"
	"github.com/kadsin/banking-system/internal/server/middlewares"
	"github.com/stretchr/testify/require"
)

func TestRequireAuth_MissingToken(t *testing.T) {
	app := fiber.New()
	app.Get("/private", middlewares.RequireAuth, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_ValidToken(t *testing.T) {
	app := fiber.New()
	app.Get("/private", middlewares.RequireAuth, func(c *fiber.Ctx) error {
		id, err := middlewares.AuthUserID(c)

		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)

		return c.SendStatus(http.StatusOK)
	})

	jwtSvc := auth.NewJWTService(config.Env.Auth.JWTSecret, config.Env.Auth.AccessTokenTTLMin)
	token, err := jwtSvc.GenerateAccessToken(uuid.New())
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	app := fiber.New()
	app.Get("/private", middlewares.RequireAuth, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
