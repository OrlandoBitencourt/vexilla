package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_GetSet(t *testing.T) {
	store, err := NewMemoryStore(1<<20, 1000, 64)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test Get non-existent
	flag, found := store.Get(ctx, "nonexistent")
	assert.False(t, found)
	assert.Nil(t, flag)

	// Test Set and Get
	testFlag := vexilla.Flag{
		ID:      1,
		Key:     "test_flag",
		Enabled: true,
	}

	ok := store.Set(ctx, "test_flag", testFlag)
	assert.True(t, ok)
	store.Wait()

	flag, found = store.Get(ctx, "test_flag")
	assert.True(t, found)
	require.NotNil(t, flag)
	assert.Equal(t, int64(1), flag.ID)
	assert.Equal(t, "test_flag", flag.Key)
	assert.True(t, flag.Enabled)
}

func TestMemoryStore_Delete(t *testing.T) {
	store, err := NewMemoryStore(1<<20, 1000, 64)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	testFlag := vexilla.Flag{Key: "test_flag"}
	store.Set(ctx, "test_flag", testFlag)
	store.Wait()

	// Verify it exists
	_, found := store.Get(ctx, "test_flag")
	assert.True(t, found)

	// Delete it
	store.Delete(ctx, "test_flag")

	// Verify it's gone
	_, found = store.Get(ctx, "test_flag")
	assert.False(t, found)
}

func TestMemoryStore_Clear(t *testing.T) {
	store, err := NewMemoryStore(1<<20, 1000, 64)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Add multiple flags
	for i := 0; i < 10; i++ {
		flag := vexilla.Flag{
			ID:  int64(i),
			Key: fmt.Sprintf("flag_%d", i),
		}
		store.Set(ctx, flag.Key, flag)
	}
	store.Wait()

	// Clear all
	store.Clear(ctx)

	// Verify all gone
	for i := 0; i < 10; i++ {
		_, found := store.Get(ctx, fmt.Sprintf("flag_%d", i))
		assert.False(t, found)
	}
}
