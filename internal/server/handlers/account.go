package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kadsin/banking-system/internal/contracts"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/kadsin/banking-system/internal/server/requests"
)

func NewAccountHandler(repo contracts.AccountRepository) *AccountHandler {
	return &AccountHandler{
		repo: repo,
	}
}

type AccountHandler struct {
	repo contracts.AccountRepository
}

func (h *AccountHandler) Create(c *fiber.Ctx) error {
	type CreateAccountRequest struct {
		Currency string `json:"currency" validate:"required"`
		Balance  int64  `json:"balance"`
	}

	body, err := requests.Prepare[CreateAccountRequest](c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	account := domain.Account{
		ID:       uuid.NewString(),
		Balance:  body.Balance,
		Currency: body.Currency,
		Status:   domain.AccountStatusActive,
	}

	account, err = h.repo.Create(account)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(account)
}

func (h *AccountHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id must be a valid UUID")
	}

	account, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, datalayer.ErrAccountNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}

		return err
	}

	return c.JSON(account)
}
