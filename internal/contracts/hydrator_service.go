package contracts

type HydratorService interface {
	Repopulate(accountID string) (int64, error)
}
