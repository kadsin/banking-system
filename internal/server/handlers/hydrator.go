package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kadsin/banking-system/internal/contracts"
)

func NewHydratorHandler(hydrator contracts.HydratorService) *HydratorHandler {
	return &HydratorHandler{hydrator: hydrator}
}

type HydratorHandler struct {
	hydrator contracts.HydratorService
}

func (h *HydratorHandler) Repopulate(c *fiber.Ctx) error {
	accountID := c.Params("account_id")
	if _, err := uuid.Parse(accountID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "account_id must be a valid UUID")
	}

	balance, err := h.hydrator.Repopulate(accountID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"account_id": accountID,
		"balance":    balance,
	})
}
