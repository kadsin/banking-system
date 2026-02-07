package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func Prepare[T comparable](c *fiber.Ctx) (body T, err error) {
	if err = c.BodyParser(&body); err != nil {
		return
	}

	if err = validator.New().Struct(&body); err != nil {
		return
	}

	return
}
