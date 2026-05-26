package queue

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueueFetchAndCommitFlow(t *testing.T) {
	q := New()

	_, err := q.Publish("tx", []byte("a"))
	require.NoError(t, err)
	_, err = q.Publish("tx", []byte("b"))
	require.NoError(t, err)
	_, err = q.Publish("tx", []byte("c"))
	require.NoError(t, err)

	batch, err := q.Fetch("tx", 2)
	require.NoError(t, err)
	require.Len(t, batch, 2)
	require.Equal(t, int64(0), batch[0].Offset)
	require.Equal(t, int64(1), batch[1].Offset)

	err = q.Commit("tx", 1)
	require.NoError(t, err)

	batch, err = q.Fetch("tx", 10)
	require.NoError(t, err)
	require.Len(t, batch, 1)
	require.Equal(t, int64(2), batch[0].Offset)
	require.Equal(t, []byte("c"), batch[0].Value)
}

func TestQueueCommitValidation(t *testing.T) {
	q := New()
	_, err := q.Publish("tx", []byte("a"))
	require.NoError(t, err)

	err = q.Commit("tx", -1)
	require.ErrorIs(t, err, ErrInvalidOffset)

	err = q.Commit("tx", 0)
	require.NoError(t, err)

	err = q.Commit("tx", 0)
	require.NoError(t, err)
}

