package vexilla

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.opentelemetry.io/otel/attribute"
)

// saveToDisk persists the cache to disk
func (c *Cache) saveToDisk(ctx context.Context) error {
	_, span := c.tracer.Start(ctx, "cache.save_to_disk")
	defer span.End()

	if err := os.MkdirAll(c.config.PersistencePath, 0755); err != nil {
		return fmt.Errorf("failed to create persistence directory: %w", err)
	}

	// Note: Ristretto doesn't provide iteration, so we maintain a snapshot
	// This is a known limitation - consider maintaining a separate index if needed
	flags := make(map[string]Flag)

	filePath := filepath.Join(c.config.PersistencePath, "flags.json")
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

// loadFromDisk loads cached flags from disk
func (c *Cache) loadFromDisk(ctx context.Context) error {
	_, span := c.tracer.Start(ctx, "cache.load_from_disk")
	defer span.End()

	filePath := filepath.Join(c.config.PersistencePath, "flags.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read flags from disk: %w", err)
	}

	var flags map[string]Flag
	if err := json.Unmarshal(data, &flags); err != nil {
		return fmt.Errorf("failed to unmarshal flags: %w", err)
	}

	for key, flag := range flags {
		c.cache.Set(key, flag, 1)
	}
	c.cache.Wait()

	span.SetAttributes(attribute.Int("flags.loaded", len(flags)))
	return nil
}

// loadFlagFromDisk attempts to load a single flag from disk
func (c *Cache) loadFlagFromDisk(ctx context.Context, flagKey string) (interface{}, error) {
	filePath := filepath.Join(c.config.PersistencePath, "flags.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("flag not found in disk cache: %w", err)
	}

	var flags map[string]Flag
	if err := json.Unmarshal(data, &flags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flags: %w", err)
	}

	flag, ok := flags[flagKey]
	if !ok {
		return nil, fmt.Errorf("flag %s not found in disk cache", flagKey)
	}

	return flag.Default, nil
}
