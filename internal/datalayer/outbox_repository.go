package datalayer

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kadsin/banking-system/internal/domain"
)

var ErrOutboxEventNotFound = errors.New("outbox event not found")

func NewOutboxRepository() *OutboxRepository {
	return &OutboxRepository{
		events: map[string]domain.OutboxEvent{},
		order:  make([]string, 0),
	}
}

type OutboxRepository struct {
	mu     sync.RWMutex
	events map[string]domain.OutboxEvent
	order  []string
}

func (r *OutboxRepository) Create(topic string, payload []byte) (domain.OutboxEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	event := domain.OutboxEvent{
		ID:        uuid.NewString(),
		Topic:     topic,
		Payload:   payload,
		Status:    domain.OutboxEventStatusPending,
		CreatedAt: time.Now().UTC(),
	}

	r.events[event.ID] = event
	r.order = append(r.order, event.ID)

	return event, nil
}

func (r *OutboxRepository) GetByID(id string) (domain.OutboxEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, ok := r.events[id]
	if !ok {
		return domain.OutboxEvent{}, ErrOutboxEventNotFound
	}

	return event, nil
}

func (r *OutboxRepository) ListPending(limit int) ([]domain.OutboxEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		return []domain.OutboxEvent{}, nil
	}

	events := make([]domain.OutboxEvent, 0, limit)
	for _, id := range r.order {
		event := r.events[id]
		if event.Status != domain.OutboxEventStatusPending {
			continue
		}

		events = append(events, event)
		if len(events) == limit {
			break
		}
	}

	return events, nil
}

func (r *OutboxRepository) MarkProcessed(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, ok := r.events[id]
	if !ok {
		return ErrOutboxEventNotFound
	}

	now := time.Now().UTC()
	event.Status = domain.OutboxEventStatusProcessed
	event.ProcessedAt = &now
	r.events[id] = event

	return nil
}
