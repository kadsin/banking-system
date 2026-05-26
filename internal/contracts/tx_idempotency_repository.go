package contracts

type TxIdempotencyRepository interface {
	Get(idempotencyKey string) (txID string, exists bool, err error)
	Set(idempotencyKey string, txID string) error
}
