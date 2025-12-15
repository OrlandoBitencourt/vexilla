package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
)

var ErrNotFound = errors.New("flag not found")

const snapshotFile = "snapshot.json"

type DiskStorage struct {
	dir     string
	metrics Metrics
	mu      sync.RWMutex
}

func NewDiskStorage(dir string) (*DiskStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &DiskStorage{dir: dir}, nil
}

func (d *DiskStorage) filePath(key string) string {
	return filepath.Join(d.dir, fmt.Sprintf("%s.json", key))
}

func (d *DiskStorage) Get(ctx context.Context, key string) (*domain.Flag, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	data, err := os.ReadFile(d.filePath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var flag domain.Flag
	if err := json.Unmarshal(data, &flag); err != nil {
		return nil, err
	}

	d.metrics.GetsKept++
	return &flag, nil
}

func (d *DiskStorage) Set(ctx context.Context, key string, flag domain.Flag, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := json.MarshalIndent(flag, "", "  ")
	if err != nil {
		return err
	}

	file := d.filePath(key)
	writeErr := os.WriteFile(file, data, 0644)
	if writeErr != nil {
		d.metrics.SetsDropped++
		return writeErr
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		d.metrics.KeysAdded++
	} else {
		d.metrics.KeysUpdated++
	}

	return nil
}

func (d *DiskStorage) Delete(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := os.Remove(d.filePath(key))
	if err != nil {
		return err
	}

	d.metrics.KeysDeleted++
	return nil
}

func (d *DiskStorage) Clear(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	entries, err := os.ReadDir(d.dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		os.Remove(filepath.Join(d.dir, e.Name()))
	}

	return nil
}

func (d *DiskStorage) List(ctx context.Context) ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entries, err := os.ReadDir(d.dir)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			keys = append(keys, entry.Name()[:len(entry.Name())-5])
		}
	}
	return keys, nil
}

func (d *DiskStorage) SaveSnapshot(ctx context.Context, snapshot map[string]domain.Flag) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	file := filepath.Join(d.dir, snapshotFile)
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	return nil
}

func (d *DiskStorage) LoadSnapshot(ctx context.Context) (map[string]domain.Flag, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	file := filepath.Join(d.dir, snapshotFile)

	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("snapshot not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	var snapshot map[string]domain.Flag
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to decode snapshot: %w", err)
	}

	return snapshot, nil
}

func (d *DiskStorage) Metrics() Metrics { return d.metrics }

func (d *DiskStorage) Close() error { return nil }
