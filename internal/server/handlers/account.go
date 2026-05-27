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

func NewAccountHandler(service contracts.AccountService) *AccountHandler {
	return &AccountHandler{
		service: service,
	}
}

type AccountHandler struct {
	service contracts.AccountService
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

	account, err = h.service.Create(account)
	if err != nil {
		return err
	}

	type CreateAccountResponse struct {
		ID       string               `json:"id"`
		Balance  int64                `json:"balance"`
		Currency string               `json:"currency"`
		Status   domain.AccountStatus `json:"status"`
	}
	return c.Status(fiber.StatusCreated).JSON(CreateAccountResponse{
		ID:       account.ID,
		Balance:  account.Balance,
		Currency: account.Currency,
		Status:   account.Status,
	})
}

func (h *AccountHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id must be a valid UUID")
	}

	account, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, datalayer.ErrAccountNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}

		return err
	}

	type GetAccountByIDResponse struct {
		ID       string               `json:"id"`
		Balance  int64                `json:"balance"`
		Currency string               `json:"currency"`
		Status   domain.AccountStatus `json:"status"`
	}
	return c.JSON(GetAccountByIDResponse{
		ID:       account.ID,
		Balance:  account.Balance,
		Currency: account.Currency,
		Status:   account.Status,
	})
}
