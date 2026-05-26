package queue

import (
	"errors"
	"sync"
)

var (
	ErrEmptyTopic       = errors.New("topic is required")
	ErrInvalidOffset    = errors.New("invalid offset")
	ErrCommitOutOfOrder = errors.New("commit offset is out of order")
)

type Message struct {
	Offset int64
	Value  []byte
}

func New() *Queue {
	return &Queue{
		messagesByTopic:  map[string][]Message{},
		committedByTopic: map[string]int64{},
	}
}

type Queue struct {
	mu               sync.RWMutex
	messagesByTopic  map[string][]Message
	committedByTopic map[string]int64
}

func (q *Queue) Publish(topic string, value []byte) (Message, error) {
	if topic == "" {
		return Message{}, ErrEmptyTopic
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	nextOffset := int64(len(q.messagesByTopic[topic]))
	message := Message{
		Offset: nextOffset,
		Value:  append([]byte(nil), value...),
	}

	q.messagesByTopic[topic] = append(q.messagesByTopic[topic], message)
	if _, ok := q.committedByTopic[topic]; !ok {
		q.committedByTopic[topic] = -1
	}

	return message, nil
}

func (q *Queue) Fetch(topic string, limit int) ([]Message, error) {
	if topic == "" {
		return nil, ErrEmptyTopic
	}
	if limit <= 0 {
		return []Message{}, nil
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	startOffset := q.committedByTopic[topic] + 1
	if startOffset < 0 {
		startOffset = 0
	}

	messages := q.messagesByTopic[topic]
	if startOffset >= int64(len(messages)) {
		return []Message{}, nil
	}

	endOffset := startOffset + int64(limit)
	if endOffset > int64(len(messages)) {
		endOffset = int64(len(messages))
	}

	result := make([]Message, 0, endOffset-startOffset)
	for _, message := range messages[startOffset:endOffset] {
		result = append(result, Message{
			Offset: message.Offset,
			Value:  append([]byte(nil), message.Value...),
		})
	}

	return result, nil
}

func (q *Queue) Commit(topic string, offset int64) error {
	if topic == "" {
		return ErrEmptyTopic
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	messages := q.messagesByTopic[topic]
	if offset < 0 || offset >= int64(len(messages)) {
		return ErrInvalidOffset
	}

	current := q.committedByTopic[topic]
	if offset < current {
		return ErrCommitOutOfOrder
	}

	q.committedByTopic[topic] = offset
	return nil
}
