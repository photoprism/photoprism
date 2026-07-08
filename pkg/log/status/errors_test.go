package status

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrCanceled(t *testing.T) {
	assert.EqualError(t, ErrCanceled, Canceled)

	wrapped := fmt.Errorf("worker stopped: %w", ErrCanceled)
	assert.True(t, errors.Is(wrapped, ErrCanceled))
	assert.False(t, errors.Is(wrapped, ErrInsufficientStorage))
}

func TestErrInsufficientStorage(t *testing.T) {
	assert.EqualError(t, ErrInsufficientStorage, InsufficientStorage)

	wrapped := fmt.Errorf("scheduler: %w (backup)", ErrInsufficientStorage)
	assert.True(t, errors.Is(wrapped, ErrInsufficientStorage))
	assert.False(t, errors.Is(wrapped, ErrCanceled))
}
