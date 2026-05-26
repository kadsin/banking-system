package domain

import "time"

type OutboxEventStatus string

const (
	OutboxEventStatusPending   OutboxEventStatus = "PENDING"
	OutboxEventStatusProcessed OutboxEventStatus = "PROCESSED"
)

type OutboxEvent struct {
	ID          string            `json:"id"`
	Topic       string            `json:"topic"`
	Payload     []byte            `json:"payload"`
	Status      OutboxEventStatus `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	ProcessedAt *time.Time        `json:"processed_at,omitempty"`
}
