package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/auth"
)

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		jwt: auth.NewJWTService(config.Env.Auth.JWTSecret, config.Env.Auth.AccessTokenTTLMin),
	}
}

type AuthHandler struct {
	jwt *auth.JWTService
}

func (a *AuthHandler) Login(c *fiber.Ctx) error {
	token, err := a.jwt.GenerateAccessToken(uuid.New())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": token,
	})
}
