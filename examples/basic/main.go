package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	log.Println("🏴 Vexilla - Basic Example")
	log.Println("============================")

	// 1. Create configuration
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.RefreshInterval = 1 * time.Minute
	config.PersistenceEnabled = true
	config.PersistencePath = "./vexilla-cache"
	config.AdminAPIEnabled = true

	// 2. Create cache instance
	log.Println("📦 Creating Vexilla cache...")
	cache, err := vexilla.New(config)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// 3. Start the cache (loads flags and starts background refresh)
	log.Println("🚀 Starting cache...")
	if err := cache.Start(); err != nil {
		log.Fatalf("Failed to start cache: %v", err)
	}
	defer cache.Stop()

	log.Println("✅ Cache started successfully!")

	// Give it a moment to load flags
	time.Sleep(500 * time.Millisecond)

	ctx := context.Background()

	// 4. Boolean flag evaluation (using helper to get actual key)
	log.Println("=== Example 1: Boolean Flag ===")
	evalCtx := vexilla.EvaluationContext{
		UserID: "user_123",
		Attributes: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
		},
	}

	// Get the actual flag key with hash
	newFeatureKey := MustGetFlagKey("new_feature")
	log.Printf("Flag key: %s\n", newFeatureKey)

	enabled := cache.EvaluateBool(ctx, newFeatureKey, evalCtx)
	log.Printf("New feature enabled: %v\n\n", enabled)
	newFeat, err := cache.Evaluate(ctx, newFeatureKey, evalCtx)
	if err != nil {
		log.Printf("New feature enabled: %v\n\n", err.Error())
	}
	log.Printf("New feature flag attackment: %v\n\n", newFeat)

	// 5. String flag evaluation
	log.Println("=== Example 2: String Flag ===")
	uiThemeKey := MustGetFlagKey("ui_theme")
	theme := cache.EvaluateString(ctx, uiThemeKey, evalCtx, "light")
	log.Printf("UI Theme: %s\n\n", theme)

	// 6. Integer flag evaluation
	log.Println("=== Example 3: Integer Flag ===")
	maxItemsKey := MustGetFlagKey("max_items")
	maxItems := cache.EvaluateInt(ctx, maxItemsKey, evalCtx, 10)
	log.Printf("Max items: %d\n\n", maxItems)

	// 7. Full evaluation with details
	log.Println("=== Example 4: Full Evaluation ===")
	brazilLaunchKey := MustGetFlagKey("brazil_launch")
	result, err := cache.Evaluate(ctx, brazilLaunchKey, vexilla.EvaluationContext{
		UserID: "user_789",
		Attributes: map[string]interface{}{
			"country":  "BR",
			"document": "1234567808",
			"age":      25,
		},
	})
	if err != nil {
		log.Printf("Evaluation error: %v\n", err)
	} else {
		log.Printf("Result: %v\n\n", result)
	}

	// 8. Cache statistics
	log.Println("=== Example 5: Cache Statistics ===")
	stats := cache.GetCacheStats()
	log.Printf("Keys Added: %d\n", stats.KeysAdded)
	log.Printf("Keys Evicted: %d\n", stats.KeysEvicted)
	log.Printf("Hit Ratio: %.2f%%\n", stats.HitRatio*100)
	log.Printf("Last Refresh: %s\n\n", stats.LastRefresh.Format(time.RFC3339))

	// 9. Manual cache invalidation
	log.Println("=== Example 6: Manual Invalidation ===")
	log.Println("Invalidating 'new_feature' flag...")
	cache.InvalidateFlag(ctx, newFeatureKey)
	log.Println("✅ Flag invalidated")

	// 10. Show all available flags
	log.Println("=== Example 7: Available Flags ===")
	allFlags, _ := GetAllFlagKeys()
	log.Println("Available flags:")
	for name, key := range allFlags {
		log.Printf("  • %-25s → %s\n", name, key)
	}
	log.Println()

	// 11. Keep running for a bit to show background refresh
	log.Println("📊 Cache is running. Background refresh will occur every minute.")
	log.Println("Press Ctrl+C to stop")

	// Simulate some flag evaluations
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 6; i++ {
		<-ticker.C
		enabled := cache.EvaluateBool(ctx, newFeatureKey, evalCtx)
		log.Printf("[%s] Feature check: %v", time.Now().Format("15:04:05"), enabled)
	}

	log.Println("\n✨ Example completed!")
}

const (
	flagrEndpoint = "http://localhost:18000"
)

// FlagKeyCache caches flag keys to avoid repeated API calls
type FlagKeyCache struct {
	keys map[string]string
	mu   sync.RWMutex
}

var globalCache = &FlagKeyCache{
	keys: make(map[string]string),
}

// GetFlagKey returns the actual flag key (with hash) for a given base name
// Example: GetFlagKey("new_feature") returns "new_feature-abc123"
func GetFlagKey(baseName string) (string, error) {
	// Check cache first
	globalCache.mu.RLock()
	if key, ok := globalCache.keys[baseName]; ok {
		globalCache.mu.RUnlock()
		return key, nil
	}
	globalCache.mu.RUnlock()

	// Fetch from Flagr
	resp, err := http.Get(flagrEndpoint + "/api/v1/flags")
	if err != nil {
		return "", fmt.Errorf("failed to fetch flags: %w", err)
	}
	defer resp.Body.Close()

	var flags []struct {
		Key         string `json:"key"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&flags); err != nil {
		return "", fmt.Errorf("failed to decode flags: %w", err)
	}

	// Build cache
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	for _, flag := range flags {
		// Description format: "FLAG:new_feature"
		if strings.HasPrefix(flag.Description, "FLAG:") {
			name := strings.TrimPrefix(flag.Description, "FLAG:")
			globalCache.keys[name] = flag.Key
		}
	}

	// Return the requested key
	if key, ok := globalCache.keys[baseName]; ok {
		return key, nil
	}

	return "", fmt.Errorf("flag '%s' not found (did you run setup-flags.go?)", baseName)
}

// MustGetFlagKey is like GetFlagKey but panics on error (for cleaner example code)
func MustGetFlagKey(baseName string) string {
	key, err := GetFlagKey(baseName)
	if err != nil {
		panic(err)
	}
	return key
}

// GetAllFlagKeys returns a map of all flag base names to their actual keys
func GetAllFlagKeys() (map[string]string, error) {
	globalCache.mu.RLock()
	if len(globalCache.keys) > 0 {
		// Return cached copy
		cached := make(map[string]string, len(globalCache.keys))
		for k, v := range globalCache.keys {
			cached[k] = v
		}
		globalCache.mu.RUnlock()
		return cached, nil
	}
	globalCache.mu.RUnlock()

	// Force refresh
	_, err := GetFlagKey("dummy")
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, err
	}

	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()

	result := make(map[string]string, len(globalCache.keys))
	for k, v := range globalCache.keys {
		result[k] = v
	}

	return result, nil
}

// ClearCache clears the flag key cache (useful if flags are recreated)
func ClearCache() {
	globalCache.mu.Lock()
	globalCache.keys = make(map[string]string)
	globalCache.mu.Unlock()
}
