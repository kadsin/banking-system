package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func ChangeUserBalance(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id": 0,
	})
}
