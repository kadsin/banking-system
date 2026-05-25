package domain

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "ACTIVE"
	AccountStatusBlocked AccountStatus = "BLOCKED"
)

type Account struct {
	ID       string        `json:"id"`
	Balance  int64         `json:"balance"`
	Currency string        `json:"currency"`
	Status   AccountStatus `json:"status"`
}
