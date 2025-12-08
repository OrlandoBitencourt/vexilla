package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskStore_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDiskStore(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Test Save
	flags := map[string]vexilla.Flag{
		"flag1": {ID: 1, Key: "flag1", Enabled: true},
		"flag2": {ID: 2, Key: "flag2", Enabled: false},
	}

	err = store.Save(ctx, flags)
	require.NoError(t, err)

	// Verify file exists
	filePath := filepath.Join(tmpDir, "flags.json")
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Test Load
	loaded, err := store.Load(ctx)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "flag1", loaded["flag1"].Key)
	assert.Equal(t, "flag2", loaded["flag2"].Key)
}

func TestDiskStore_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewDiskStore(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Load from non-existent file should return empty map
	flags, err := store.Load(ctx)
	require.NoError(t, err)
	assert.Empty(t, flags)
}

func TestDiskStore_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "nested", "path")

	store, err := NewDiskStore(storePath)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(storePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
