package domain

import "time"

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

type Transaction struct {
	ID             string            `json:"id"`
	FromAccountID  string            `json:"from_account"`
	ToAccountID    string            `json:"to_account"`
	Amount         int64             `json:"amount"`
	Status         TransactionStatus `json:"status"`
	IdempotencyKey string            `json:"idempotency_key"`
	Timestamp      time.Time         `json:"timestamp"`
}
