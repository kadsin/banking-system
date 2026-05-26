package contracts

type BalanceRepository interface {
	Adjust(accountID string, delta int64) error
	Get(accountID string) (int64, error)
}
