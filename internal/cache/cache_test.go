package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheSetAndGet(t *testing.T) {
	c := New()

	err := c.Set("a", "1")
	require.NoError(t, err)

	value, err := c.Get("a")
	require.NoError(t, err)
	require.Equal(t, "1", value)
}

func TestCacheValidation(t *testing.T) {
	c := New()

	_, err := c.Get("")
	require.ErrorIs(t, err, ErrEmptyKey)

	err = c.Set("", "v")
	require.ErrorIs(t, err, ErrEmptyKey)
}
