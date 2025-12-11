package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	log.Println("🔍 Vexilla Debug Tool - Enhanced")
	log.Println("=================================")

	flagrURL := "http://localhost:18000"

	// 1. Test Flagr API directly
	log.Println("\n📡 Step 1: Testing direct Flagr API call...")
	resp, err := http.Get(flagrURL + "/api/v1/flags")
	if err != nil {
		log.Fatalf("❌ Failed to connect to Flagr: %v", err)
	}
	defer resp.Body.Close()

	var flagsRaw []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flagsRaw); err != nil {
		log.Fatalf("❌ Failed to decode: %v", err)
	}

	log.Printf("✅ Found %d flags in Flagr\n", len(flagsRaw))
	for _, flag := range flagsRaw {
		if key, ok := flag["key"].(string); ok {
			log.Printf("   • %s (enabled: %v)\n", key, flag["enabled"])
		}
	}

	// 2. Test direct evaluation with Flagr (correct endpoint)
	log.Println("\n🔄 Step 2: Testing direct Flagr evaluation API...")
	flagKey := "new_feature-6a65f91d"

	evalReq := map[string]interface{}{
		"entityID": "user_123",
		"entityContext": map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
		},
	}

	body, _ := json.Marshal(evalReq)

	// Corrected endpoint - note the /evaluation without /flags/ prefix
	evalURL := fmt.Sprintf("%s/api/v1/evaluation", flagrURL)
	log.Printf("   POST %s\n", evalURL)
	log.Printf("   Body: %s\n", string(body))

	evalResp, err := http.Post(evalURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("❌ Direct eval failed: %v\n", err)
	} else {
		defer evalResp.Body.Close()
		bodyBytes, _ := io.ReadAll(evalResp.Body)
		log.Printf("   Status: %d\n", evalResp.StatusCode)
		log.Printf("   Response: %s\n", string(bodyBytes))

		if evalResp.StatusCode == 200 {
			var evalResult map[string]interface{}
			json.Unmarshal(bodyBytes, &evalResult)
			log.Printf("✅ Direct eval result: %+v\n", evalResult)
		}
	}

	// 3. Try batch evaluation endpoint
	log.Println("\n🔄 Step 3: Testing batch evaluation endpoint...")
	batchURL := fmt.Sprintf("%s/api/v1/evaluation/batch", flagrURL)
	batchReq := map[string]interface{}{
		"entities": []map[string]interface{}{
			{
				"entityID": "user_123",
				"entityContext": map[string]interface{}{
					"country": "BR",
					"tier":    "premium",
				},
			},
		},
		"flagKeys": []string{flagKey},
	}

	batchBody, _ := json.Marshal(batchReq)
	log.Printf("   POST %s\n", batchURL)

	batchResp, err := http.Post(batchURL, "application/json", bytes.NewReader(batchBody))
	if err != nil {
		log.Printf("❌ Batch eval failed: %v\n", err)
	} else {
		defer batchResp.Body.Close()
		bodyBytes, _ := io.ReadAll(batchResp.Body)
		log.Printf("   Status: %d\n", batchResp.StatusCode)
		log.Printf("   Response: %s\n", string(bodyBytes))
	}

	// 4. Test Vexilla cache with verbose logging
	log.Println("\n📦 Step 4: Creating Vexilla cache...")
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = flagrURL
	config.RefreshInterval = 1 * time.Minute
	config.PersistenceEnabled = false // Disable persistence for debugging

	cache, err := vexilla.New(config)
	if err != nil {
		log.Fatalf("❌ Failed to create cache: %v", err)
	}

	log.Println("🚀 Step 5: Starting cache...")
	if err := cache.Start(); err != nil {
		log.Fatalf("❌ Failed to start cache: %v", err)
	}
	defer cache.Stop()

	// Wait longer for loading
	log.Println("⏳ Waiting for flags to load...")
	time.Sleep(3 * time.Second)

	stats := cache.GetCacheStats()
	log.Printf("📊 Cache statistics:\n")
	log.Printf("   Keys Added: %d\n", stats.KeysAdded)
	log.Printf("   Keys Evicted: %d\n", stats.KeysEvicted)
	log.Printf("   Hit Ratio: %.2f%%\n", stats.HitRatio*100)
	log.Printf("   Last Refresh: %v\n", stats.LastRefresh)

	if stats.KeysAdded == 0 {
		log.Println("⚠️  WARNING: No flags were loaded into cache!")
		log.Println("   This suggests a problem with Vexilla's flag loading logic.")
	}

	// 5. Test flag evaluation
	log.Println("\n🎯 Step 6: Testing flag evaluation with Vexilla...")
	ctx := context.Background()

	evalCtx := vexilla.EvaluationContext{
		UserID: "user_123",
		Attributes: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
		},
	}

	log.Printf("   Flag Key: %s\n", flagKey)
	log.Printf("   User ID: %s\n", evalCtx.UserID)

	result, err := cache.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		log.Printf("❌ Evaluation failed: %v\n", err)
	} else {
		log.Printf("✅ Evaluation result: %v (type: %T)\n", result, result)
	}

	boolResult := cache.EvaluateBool(ctx, flagKey, evalCtx)
	log.Printf("   EvaluateBool result: %v\n", boolResult)

	// 6. Check if flag exists in cache
	log.Println("\n🔍 Step 7: Checking cache internals...")
	// Note: This might need to be adjusted based on Vexilla's internal API
	// You may need to add a debug method to Vexilla to inspect cache contents

	log.Print("\n" + "=")
	log.Println("📋 DIAGNOSIS:")
	log.Print("=")

	if stats.KeysAdded == 0 {
		log.Println("❌ PROBLEM: Vexilla is NOT loading flags from Flagr")
		log.Println("   Possible causes:")
		log.Println("   1. Bug in Vexilla's Start() or loadFlags() method")
		log.Println("   2. Incorrect API endpoint being called")
		log.Println("   3. Response format mismatch")
		log.Println()
		log.Println("🔧 RECOMMENDED ACTIONS:")
		log.Println("   1. Check Vexilla source code for the loadFlags() implementation")
		log.Println("   2. Add debug logs to Vexilla's flag loading logic")
		log.Println("   3. Verify the API endpoint Vexilla is calling")
		log.Println("   4. Check if Vexilla expects a specific response format")
	} else {
		log.Println("✅ Flags loaded successfully")
	}

	log.Println("\n✨ Debug complete!")
}
