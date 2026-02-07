package middlewares

import (
	"errors"
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/kadsin/banking-system/config"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
	}

	if statusCode >= fiber.StatusInternalServerError {
		log.Printf("Error: %+v\n%s\n", err, debug.Stack())

		if !config.Env.App.Debug {
			err = errors.New("server error")
		}
	}

	return c.Status(statusCode).SendString(err.Error())
}
