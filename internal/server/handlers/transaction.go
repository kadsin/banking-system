package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/server/requests"
	"github.com/kadsin/banking-system/internal/service"
)

func NewTransactionHandler(txRepo contracts.OlapRepository, transfer contracts.TransferService) *TransactionHandler {
	return &TransactionHandler{
		txRepo:   txRepo,
		transfer: transfer,
	}
}

type TransactionHandler struct {
	txRepo   contracts.OlapRepository
	transfer contracts.TransferService
}

func (h *TransactionHandler) Transfer(c *fiber.Ctx) error {
	type TransferRequest struct {
		FromAccountID  string `json:"from_account" validate:"required,uuid4"`
		ToAccountID    string `json:"to_account" validate:"required,uuid4"`
		Amount         int64  `json:"amount" validate:"required,gt=0"`
		IdempotencyKey string `json:"idempotency_key" validate:"required"`
	}

	body, err := requests.Prepare[TransferRequest](c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if body.FromAccountID == body.ToAccountID {
		return fiber.NewError(fiber.StatusBadRequest, "The origin and destination should not be same")
	}

	tx, err := h.transfer.Transfer(contracts.TransferInput{
		FromAccountID:  body.FromAccountID,
		ToAccountID:    body.ToAccountID,
		Amount:         body.Amount,
		IdempotencyKey: body.IdempotencyKey,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientFunds):
			return fiber.NewError(fiber.StatusConflict, err.Error())
		case errors.Is(err, service.ErrAccountBlocked):
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		case errors.Is(err, datalayer.ErrAccountNotFound):
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		default:
			return err
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(tx)
}

func (h *TransactionHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id must be a valid UUID")
	}

	tx, err := h.txRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, datalayer.ErrTransactionNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}

		return err
	}

	return c.JSON(tx)
}

func (h *TransactionHandler) History(c *fiber.Ctx) error {
	accountID := c.Params("account_id")
	if _, err := uuid.Parse(accountID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "account_id must be a valid UUID")
	}

	history, err := h.txRepo.ListByAccountID(accountID)
	if err != nil {
		return err
	}

	return c.JSON(history)
}
