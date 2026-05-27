package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kadsin/banking-system/bootstrap"
	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/cache"
	"github.com/kadsin/banking-system/internal/cdc"
	"github.com/kadsin/banking-system/internal/datalayer"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/kadsin/banking-system/internal/queue"
	"github.com/kadsin/banking-system/internal/server"
	"github.com/kadsin/banking-system/internal/service"
	"github.com/stretchr/testify/require"
)

type envelope[T any] struct {
	Data   T `json:"data"`
	Errors []struct {
		Detail string `json:"detail"`
	} `json:"errors"`
}

type testAPI struct {
	app *fiber.App
	cdc *cdc.App
}

func TestIntegration_AccountCreateAndGetByID(t *testing.T) {
	api := newTestAPI(t)

	token := login(t, api.app)

	createBody := map[string]any{
		"currency": "USD",
		"balance":  int64(1200),
	}

	createResp := doJSON(t, api.app, "POST", "/api/accounts", token, createBody)
	require.Equal(t, http.StatusCreated, createResp.Code, createResp.Body.String())

	var created envelope[domain.Account]
	require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))
	require.Empty(t, created.Errors)
	require.Equal(t, int64(1200), created.Data.Balance)
	require.Equal(t, domain.AccountStatusActive, created.Data.Status)
	require.NotEmpty(t, created.Data.ID)

	getResp := doJSON(t, api.app, "GET", "/api/accounts/"+created.Data.ID, token, nil)
	require.Equal(t, http.StatusOK, getResp.Code)

	var fetched envelope[domain.Account]
	require.NoError(t, json.Unmarshal(getResp.Body.Bytes(), &fetched))
	require.Empty(t, fetched.Errors)
	require.Equal(t, created.Data.ID, fetched.Data.ID)
	require.Equal(t, int64(1200), fetched.Data.Balance)
}

func TestIntegration_TransferThenGetByIDAndHistory(t *testing.T) {
	api := newTestAPI(t)
	token := login(t, api.app)

	from := createAccount(t, api.app, token, 5000)
	to := createAccount(t, api.app, token, 1000)

	idempotencyKey := "itx-" + uuid.NewString()
	transferBody := map[string]any{
		"from_account":    from.ID,
		"to_account":      to.ID,
		"amount":          int64(700),
		"idempotency_key": idempotencyKey,
	}

	transferResp := doJSON(t, api.app, "POST", "/api/transactions/transfer", token, transferBody)
	require.Equal(t, http.StatusAccepted, transferResp.Code)

	var transferred envelope[domain.Transaction]
	require.NoError(t, json.Unmarshal(transferResp.Body.Bytes(), &transferred))
	require.Empty(t, transferred.Errors)
	require.Equal(t, from.ID, transferred.Data.FromAccountID)
	require.Equal(t, to.ID, transferred.Data.ToAccountID)
	require.Equal(t, int64(700), transferred.Data.Amount)
	require.NotEmpty(t, transferred.Data.ID)

	require.NoError(t, waitForProjection(api, token, transferred.Data.ID))

	getTxResp := doJSON(t, api.app, "GET", "/api/transactions/"+transferred.Data.ID, token, nil)
	require.Equal(t, http.StatusOK, getTxResp.Code)

	var tx envelope[domain.Transaction]
	require.NoError(t, json.Unmarshal(getTxResp.Body.Bytes(), &tx))
	require.Empty(t, tx.Errors)
	require.Equal(t, transferred.Data.ID, tx.Data.ID)

	historyResp := doJSON(t, api.app, "GET", "/api/transactions/"+from.ID+"/history", token, nil)
	require.Equal(t, http.StatusOK, historyResp.Code)

	var history envelope[[]domain.Transaction]
	require.NoError(t, json.Unmarshal(historyResp.Body.Bytes(), &history))
	require.Empty(t, history.Errors)
	require.NotEmpty(t, history.Data)

	found := false
	for _, item := range history.Data {
		if item.ID == transferred.Data.ID {
			found = true
			break
		}
	}
	require.True(t, found, "transferred transaction should appear in source account history")
}

func newTestAPI(t *testing.T) testAPI {
	t.Helper()

	q := queue.New()
	ledgerRepo := datalayer.NewLedgerRepository(cache.New())
	balanceService := service.NewBalanceService(
		ledgerRepo,
		service.NewHydratorService(ledgerRepo, datalayer.NewHydratorRepository(cache.New()), q),
	)
	accountRepo := datalayer.NewAccountRepository()
	olapRepo := datalayer.NewOlapRepository(q)
	outboxRepo := datalayer.NewOutboxRepository()
	idempotencyRepo := datalayer.NewTxIdempotencyRepository(cache.New())

	require.NoError(t, ledgerRepo.Set(domain.SystemAccountID, config.Env.Treasury.InitialBalance))
	_, err := accountRepo.Create(domain.Account{
		ID:       domain.SystemAccountID,
		Balance:  config.Env.Treasury.InitialBalance,
		Currency: "USD",
		Status:   domain.AccountStatusActive,
	})
	require.NoError(t, err)

	transferService := service.NewTransferService(accountRepo, olapRepo, balanceService, outboxRepo, idempotencyRepo)
	transactionService := service.NewTransactionService(olapRepo)
	accountService := service.NewAccountService(accountRepo, balanceService, transferService)

	app := bootstrap.SetupFiberApp(&server.Dependencies{
		Accounts:     accountService,
		Transactions: transactionService,
		Transferer:   transferService,
	})

	return testAPI{
		app: app,
		cdc: cdc.New(outboxRepo, q),
	}
}

func login(t *testing.T, app *fiber.App) string {
	t.Helper()

	resp := doJSON(t, app, "POST", "/api/login", "", map[string]any{})
	require.Equal(t, http.StatusCreated, resp.Code)

	var body envelope[map[string]string]
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	require.Empty(t, body.Errors)
	token := body.Data["token"]
	require.NotEmpty(t, token)

	return token
}

func createAccount(t *testing.T, app *fiber.App, token string, balance int64) domain.Account {
	t.Helper()

	resp := doJSON(t, app, "POST", "/api/accounts", token, map[string]any{
		"currency": "USD",
		"balance":  balance,
	})
	require.Equal(t, http.StatusCreated, resp.Code, resp.Body.String())

	var body envelope[domain.Account]
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	require.Empty(t, body.Errors)

	return body.Data
}

func waitForProjection(api testAPI, token, txID string) error {
	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		if _, err := api.cdc.PullAndPush(100); err != nil {
			return err
		}

		resp := doJSON(nil, api.app, "GET", "/api/transactions/"+txID, token, nil)
		if resp.Code == http.StatusOK {
			return nil
		}

		time.Sleep(80 * time.Millisecond)
	}

	return context.DeadlineExceeded
}

func doJSON(t *testing.T, app *fiber.App, method, path, token string, body any) *httptest.ResponseRecorder {
	if t != nil {
		t.Helper()
	}

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if t != nil {
			require.NoError(t, err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := app.Test(req, -1)
	if t != nil {
		require.NoError(t, err)
	}
	if err != nil {
		return httptest.NewRecorder()
	}

	rr := httptest.NewRecorder()
	rr.Code = resp.StatusCode
	defer resp.Body.Close()
	_, _ = rr.Body.ReadFrom(resp.Body)
	return rr
}
