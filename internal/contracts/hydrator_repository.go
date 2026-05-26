package contracts

type HydratorRepository interface {
	SetSnapshot(accountID string, snapshot HydratorSnapshot) error
	GetSnapshot(accountID string) (HydratorSnapshot, error)
}

type HydratorSnapshot struct {
	Balance     int64
	QueueOffset int64
}
