package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/auth"
)

const UserIdLocalsKey = "auth_user_id"

func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
	}

	jwtSvc := auth.NewJWTService(config.Env.Auth.JWTSecret, config.Env.Auth.AccessTokenTTLMin)
	userID, err := jwtSvc.ParseAccessToken(
		strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer ")),
	)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}

	c.Locals(UserIdLocalsKey, userID)

	return c.Next()
}

func AuthUserID(c *fiber.Ctx) (uuid.UUID, error) {
	id, ok := c.Locals(UserIdLocalsKey).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	return id, nil
}
