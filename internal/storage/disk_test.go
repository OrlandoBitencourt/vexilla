package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskStorage_SetAndGet(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	flag := domain.Flag{ID: 1, Key: "test-flag", Enabled: true}

	err = ds.Set(ctx, "test-flag", flag, time.Minute)
	require.NoError(t, err)

	out, err := ds.Get(ctx, "test-flag")
	require.NoError(t, err)
	assert.Equal(t, flag.Key, out.Key)
}

func TestDiskStorage_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	_, err = ds.Get(ctx, "missing")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestDiskStorage_Delete(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	flag := domain.Flag{ID: 2, Key: "delete-me"}

	err = ds.Set(ctx, "delete-me", flag, time.Minute)
	require.NoError(t, err)

	err = ds.Delete(ctx, "delete-me")
	require.NoError(t, err)

	_, err = ds.Get(ctx, "delete-me")
	assert.Equal(t, ErrNotFound, err)
}

func TestDiskStorage_Clear(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	ds.Set(ctx, "a", domain.Flag{Key: "a"}, time.Minute)
	ds.Set(ctx, "b", domain.Flag{Key: "b"}, time.Minute)

	err = ds.Clear(ctx)
	require.NoError(t, err)

	_, err = ds.Get(ctx, "a")
	assert.Error(t, err)
}

func TestDiskStorage_List(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	ds.Set(ctx, "x", domain.Flag{Key: "x"}, time.Minute)
	ds.Set(ctx, "y", domain.Flag{Key: "y"}, time.Minute)

	keys, err := ds.List(ctx)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"x", "y"}, keys)
}

func TestDiskStorage_SaveSnapshot_LoadSnapshot(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	snap := map[string]domain.Flag{
		"flag1": {ID: 1, Key: "flag1", Enabled: true},
		"flag2": {ID: 2, Key: "flag2", Enabled: false},
	}

	err = ds.SaveSnapshot(ctx, snap)
	require.NoError(t, err)

	loaded, err := ds.LoadSnapshot(ctx)
	require.NoError(t, err)

	assert.Equal(t, snap["flag1"].Key, loaded["flag1"].Key)
	assert.Equal(t, snap["flag2"].Enabled, loaded["flag2"].Enabled)
}

func TestDiskStorage_LoadSnapshot_NotFound(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	_, err = ds.LoadSnapshot(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot not found")
}

func TestDiskStorage_LoadSnapshot_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	file := filepath.Join(dir, snapshotFile)
	os.WriteFile(file, []byte("invalid-json"), 0644)

	_, err = ds.LoadSnapshot(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode")
}

func TestDiskStorage_Metrics(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ds, err := NewDiskStorage(dir)
	require.NoError(t, err)

	flag := domain.Flag{Key: "m"}

	err = ds.Set(ctx, "m", flag, time.Minute)
	require.NoError(t, err)

	_, err = ds.Get(ctx, "m")
	require.NoError(t, err)

	err = ds.Delete(ctx, "m")
	require.NoError(t, err)

	m := ds.Metrics()

	assert.Equal(t, uint64(1), m.GetsKept)
	assert.Equal(t, uint64(0), m.GetsDropped)
	assert.Equal(t, uint64(1), m.KeysDeleted)
}
