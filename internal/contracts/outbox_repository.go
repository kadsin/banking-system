package contracts

import "github.com/kadsin/banking-system/internal/domain"

type OutboxRepository interface {
	Create(topic string, payload []byte) (domain.OutboxEvent, error)
	GetByID(id string) (domain.OutboxEvent, error)
	ListPending(limit int) ([]domain.OutboxEvent, error)
	MarkProcessed(id string) error
}
