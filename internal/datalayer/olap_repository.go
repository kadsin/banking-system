package datalayer

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/kadsin/banking-system/config"
	"github.com/kadsin/banking-system/internal/domain"
	"github.com/kadsin/banking-system/internal/queue"
)

var ErrTransactionNotFound = errors.New("transaction not found")

func NewOlapRepository(q *queue.Queue) *OlapRepository {
	r := &OlapRepository{
		transactions:  map[string]domain.Transaction{},
		offsetByTopic: map[string]int64{},
	}

	if q != nil {
		go r.consumeQueueForever(q)
	}

	return r
}

type OlapRepository struct {
	mu            sync.RWMutex
	transactions  map[string]domain.Transaction
	offsetByTopic map[string]int64
}

func (r *OlapRepository) consumeQueueForever(q *queue.Queue) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if err := r.pullFromQueue(q, config.Env.Topics.Transactions, 100); err != nil {
			log.Printf("olap repository consumer error: %v", err)
		}
	}
}

func (r *OlapRepository) pullFromQueue(q *queue.Queue, topic string, limit int) error {
	r.mu.RLock()
	lastOffset, ok := r.offsetByTopic[topic]
	r.mu.RUnlock()
	if !ok {
		lastOffset = -1
	}

	messages, err := q.FetchFromOffset(topic, lastOffset+1, limit)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	r.mu.Lock()
	for _, message := range messages {
		var tx domain.Transaction
		if decodeErr := json.Unmarshal(message.Value, &tx); decodeErr != nil {
			r.mu.Unlock()
			return decodeErr
		}

		r.transactions[tx.ID] = tx
		r.offsetByTopic[topic] = message.Offset
	}
	r.mu.Unlock()

	return nil
}

func (r *OlapRepository) Create(transaction domain.Transaction) (domain.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[transaction.ID] = transaction
	return transaction, nil
}

func (r *OlapRepository) GetByID(id string) (domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transaction, ok := r.transactions[id]
	if !ok {
		return domain.Transaction{}, ErrTransactionNotFound
	}

	return transaction, nil
}

func (r *OlapRepository) ListByAccountID(accountID string) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	history := make([]domain.Transaction, 0)
	for _, transaction := range r.transactions {
		if transaction.FromAccountID == accountID || transaction.ToAccountID == accountID {
			history = append(history, transaction)
		}
	}

	return history, nil
}

func (r *OlapRepository) UpdateStatus(id string, status domain.TransactionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	transaction, ok := r.transactions[id]
	if !ok {
		return ErrTransactionNotFound
	}

	transaction.Status = status
	r.transactions[id] = transaction

	return nil
}
