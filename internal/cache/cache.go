package cache

import (
	"errors"
	"sync"
)

var (
	ErrEmptyKey    = errors.New("key is required")
	ErrKeyNotFound = errors.New("key not found")
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func New() *Cache {
	return &Cache{
		items: map[string]string{},
	}
}

func (c *Cache) Set(key, value string) error {
	if key == "" {
		return ErrEmptyKey
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
	return nil
}

func (c *Cache) Get(key string) (string, error) {
	if key == "" {
		return "", ErrEmptyKey
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.items[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	return value, nil
}
