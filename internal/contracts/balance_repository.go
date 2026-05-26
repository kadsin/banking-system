package contracts

type BalanceRepository interface {
	Set(accountID string, balance int64) error
	Adjust(accountID string, delta int64) error
	Get(accountID string) (int64, error)
}
