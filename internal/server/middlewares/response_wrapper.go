package middlewares

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type jsonApiResponse struct {
	Data   any                   `json:"data,omitempty"`
	Errors []*jsonApiErrorObject `json:"errors,omitempty"`
}

type jsonApiErrorObject struct {
	Detail string `json:"detail,omitempty"`
}

func ResponseWrapper(c *fiber.Ctx) error {
	err := c.Next()
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			statusCode = fiberErr.Code
		}

		c.Status(statusCode)

		return c.JSON(jsonApiResponse{
			Errors: []*jsonApiErrorObject{
				{Detail: err.Error()},
			},
		})
	}

	contentType := string(c.Response().Header.ContentType())
	if strings.HasPrefix(contentType, "application/json") {
		return c.JSON(jsonApiResponse{
			Data: json.RawMessage(c.Response().Body()),
		})
	}

	return nil
}
