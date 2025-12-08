package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DiskStore handles disk persistence for flags
type DiskStore struct {
	path   string
	tracer trace.Tracer
}

// NewDiskStore creates a new disk store
func NewDiskStore(path string) (*DiskStore, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &DiskStore{
		path:   path,
		tracer: otel.Tracer("vexilla.storage.disk"),
	}, nil
}

// Save saves flags to disk
func (d *DiskStore) Save(ctx context.Context, flags map[string]vexilla.Flag) error {
	ctx, span := d.tracer.Start(ctx, "disk.save")
	defer span.End()

	filePath := filepath.Join(d.path, "flags.json")

	data, err := json.MarshalIndent(flags, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal flags: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write flags to disk: %w", err)
	}

	span.SetAttributes(attribute.Int("flags.count", len(flags)))
	return nil
}

// Load loads flags from disk
func (d *DiskStore) Load(ctx context.Context) (map[string]vexilla.Flag, error) {
	ctx, span := d.tracer.Start(ctx, "disk.load")
	defer span.End()

	filePath := filepath.Join(d.path, "flags.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]vexilla.Flag), nil
		}
		return nil, fmt.Errorf("failed to read flags from disk: %w", err)
	}

	var flags map[string]vexilla.Flag
	if err := json.Unmarshal(data, &flags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flags: %w", err)
	}

	span.SetAttributes(attribute.Int("flags.loaded", len(flags)))
	return flags, nil
}
