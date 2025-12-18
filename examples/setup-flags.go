// examples/setup-flags.go
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	flagrEndpoint = "http://localhost:18000"
)

type FlagrFlag struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Enabled     bool     `json:"enabled"`
}

type FlagrSegment struct {
	Description    string            `json:"description"`
	RolloutPercent int               `json:"rolloutPercent"`
	Constraints    []FlagrConstraint `json:"constraints"`
}

type FlagrConstraint struct {
	Property string      `json:"property"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type FlagrVariant struct {
	Key        string      `json:"key"`
	Attachment interface{} `json:"attachment"`
}

type FlagrDistribution struct {
	VariantKey string `json:"variantKey"`
	VariantID  int64  `json:"variantID"`
	Percentage int    `json:"percent"`
	Rank       int64  `json:"rank"`
}

type FlagrDistribuitions struct {
	Distribuitions []FlagrDistribution `json:"distributions"`
}

func randomKey(base string) string {
	return base
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%s-%x", base, b)
}

func main() {
	fmt.Println("🏴 Vexilla - Flagr Setup Script")
	fmt.Println("=================================")

	// Check if Flagr is running
	if !checkFlagrHealth() {
		fmt.Println("❌ Flagr is not running on", flagrEndpoint)
		fmt.Println("💡 Start Flagr with: docker run -it -p 18000:18000 ghcr.io/openflagr/flagr")
		os.Exit(1)
	}

	fmt.Println("✅ Flagr is running")

	// Delete existing flags (clean slate)
	fmt.Println("🧹 Cleaning up existing flags...")

	deleteAllFlags()
	fmt.Println()

	// Create all flags needed for examples
	flags := []struct {
		name string
		fn   func() error
	}{
		{"new_feature", createNewFeatureFlag},
		{"dark_mode", createDarkModeFlag},
		{"ui_theme", createUIThemeFlag},
		{"max_items", createMaxItemsFlag},
		{"premium_features", createPremiumFeaturesFlag},
		{"beta_access", createBetaAccessFlag},
		{"button_color_test", createButtonColorTestFlag},
		{"pricing_layout", createPricingLayoutFlag},
		{"gradual_rollout_30", createGradualRollout30Flag},
		{"brazil_launch", createBrazilLaunchFlag},
	}

	successCount := 0
	for _, flag := range flags {
		fmt.Printf("📝 Creating flag: %-25s ", flag.name)
		if err := flag.fn(); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Println("✅")
			successCount++
		}
		time.Sleep(200 * time.Millisecond) // Rate limiting
	}

	fmt.Printf("\n🎉 Setup complete! %d/%d flags created successfully\n", successCount, len(flags))
	fmt.Println("🚀 You can now run the examples!")
	fmt.Println("💡 Access Flagr UI at:", flagrEndpoint)
}

func touchGrass() {
	time.Sleep(60 * time.Millisecond)
}

func checkFlagrHealth() bool {
	resp, err := http.Get(flagrEndpoint + "/api/v1/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func deleteAllFlags() {
	resp, err := http.Get(flagrEndpoint + "/api/v1/flags")
	if err != nil {
		fmt.Printf("deleteAllFlags() %s \n", err.Error())
		return
	}
	defer resp.Body.Close()

	var flags []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&flags)

	for _, flag := range flags {
		id := int(flag["id"].(float64))

		url := fmt.Sprintf("%s/api/v1/flags/%d", flagrEndpoint, id)
		req, _ := http.NewRequest("DELETE", url, nil)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("deleteAllFlags() %s \n", err.Error())
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("deleteAllFlags() FAILED (%d): %s\n", resp.StatusCode, string(body))
		}
	}
}

func enableFlag(flagID int64) error {
	body := []byte(`{"enabled":true}`)

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/flags/%d/enabled", flagrEndpoint, flagID),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// ⚠️ Se seu Flagr usa auth básica (admin:senha)
	req.SetBasicAuth("admin", "senha_super_segura")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("enableFlag() status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ============================================
// FLAG CREATION FUNCTIONS
// ============================================

func createNewFeatureFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("new_feature"),
		Description: "FLAG:new_feature",
		Enabled:     true,
		Tags:        []string{},
	})
	if err != nil {
		return err
	}
	touchGrass()

	variantKey := "enabled"
	variantID, err := createVariant(flagID, FlagrVariant{
		Key:        variantKey,
		Attachment: map[string]interface{}{"value": true},
	})
	if err != nil {
		return err
	}
	touchGrass()

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "All users",
		RolloutPercent: 100,
	})
	if err != nil {
		return err
	}
	touchGrass()

	if err := createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: variantKey,
				VariantID:  variantID,
				Percentage: 100,
			},
		},
	}); err != nil {
		return err
	}

	return enableFlag(flagID)
}

func createDarkModeFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("dark_mode"),
		Description: "FLAG:dark_mode",
		Enabled:     true,
		Tags:        []string{},
	})

	if err != nil {
		return err
	}

	variantKey := "enabled"
	variantID, err := createVariant(flagID, FlagrVariant{
		Key:        variantKey,
		Attachment: map[string]interface{}{"value": true},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "All users",
		RolloutPercent: 100,
	})
	if err != nil {
		return err
	}
	err = createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: variantKey,
				VariantID:  variantID,
				Percentage: 100,
			},
		},
	})
	if err != nil {
		return err
	}
	return enableFlag(flagID)
}

func createUIThemeFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("ui_theme"),
		Description: "FLAG:ui_theme",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	// Create variants
	darkID, err := createVariant(flagID, FlagrVariant{
		Key:        "dark",
		Attachment: map[string]interface{}{"theme": "dark"},
	})
	if err != nil {
		return err
	}

	lightID, err := createVariant(flagID, FlagrVariant{
		Key:        "light",
		Attachment: map[string]interface{}{"theme": "light"},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "50/50 split",
		RolloutPercent: 100,
	})
	if err != nil {
		return err
	}

	dist := FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: "dark",
				VariantID:  darkID,
				Percentage: 50,
			},
			{
				VariantKey: "light",
				VariantID:  lightID,
				Percentage: 50,
			},
		},
	}
	err = createDistribution(flagID, segmentID, dist)
	if err != nil {
		return err
	}
	return enableFlag(flagID)
}

func createMaxItemsFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("max_items"),
		Description: "FLAG:max_items",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	variantKey := "100"
	variantID, err := createVariant(flagID, FlagrVariant{
		Key:        variantKey,
		Attachment: map[string]interface{}{"value": 100},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "All users",
		RolloutPercent: 100,
	})
	if err != nil {
		return err
	}

	err = createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: variantKey,
				VariantID:  variantID,
				Percentage: 100,
			},
		},
	})
	if err != nil {
		return err
	}

	return enableFlag(flagID)
}

func createPremiumFeaturesFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("premium_features"),
		Description: "FLAG:premium_features",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	variantKey := "enabled"
	variantID, err := createVariant(flagID, FlagrVariant{
		Key:        variantKey,
		Attachment: map[string]interface{}{"value": true},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "Premium users only",
		RolloutPercent: 100,
		Constraints: []FlagrConstraint{
			{
				Property: "tier",
				Operator: "EQ",
				Value:    "premium",
			},
		},
	})
	if err != nil {
		return err
	}

	err = createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: variantKey,
				VariantID:  variantID,
				Percentage: 100,
			},
		},
	})
	if err != nil {
		return err
	}
	return enableFlag(flagID)
}

func createBetaAccessFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("beta_access"),
		Description: "FLAG:beta_access",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	variantKey := "enabled"
	variantID, err := createVariant(flagID, FlagrVariant{
		Key:        variantKey,
		Attachment: map[string]interface{}{"value": true},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "Beta users",
		RolloutPercent: 100,
	})
	if err != nil {
		return err
	}

	err = createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: variantKey,
				VariantID:  variantID,
				Percentage: 100,
			},
		},
	})
	if err != nil {
		return err
	}

	return enableFlag(flagID)
}

func createButtonColorTestFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("button_color_test"),
		Description: "FLAG:button_color_test",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	blueID, err := createVariant(flagID, FlagrVariant{
		Key:        "blue",
		Attachment: map[string]interface{}{"color": "blue"},
	})
	if err != nil {
		return err
	}

	redID, err := createVariant(flagID, FlagrVariant{
		Key:        "red",
		Attachment: map[string]interface{}{"color": "red"},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "50/50 A/B test",
		RolloutPercent: 100,
	})
	if err != nil {
		return err
	}
	err = createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: "blue",
				VariantID:  blueID,
				Percentage: 50,
			},
			{
				VariantKey: "red",
				VariantID:  redID,
				Percentage: 50,
			},
		},
	})
	if err != nil {
		return err
	}

	return enableFlag(flagID)
}

func createPricingLayoutFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("pricing_layout"),
		Description: "FLAG:pricing_layout",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	variants := []string{"standard", "compact", "detailed"}
	variantIDs := make([]int64, 0)

	for _, variant := range variants {
		vid, err := createVariant(flagID, FlagrVariant{
			Key:        variant,
			Attachment: map[string]interface{}{"layout": variant},
		})
		if err != nil {
			return err
		}
		variantIDs = append(variantIDs, vid)
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "33/33/33 split",
		RolloutPercent: 100,
	})
	if err != nil {
		return err
	}

	err = createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: "standard",
				VariantID:  variantIDs[0],
				Percentage: 33,
			},
			{
				VariantKey: "compact",
				VariantID:  variantIDs[1],
				Percentage: 33,
			},
			{
				VariantKey: "detailed",
				VariantID:  variantIDs[2],
				Percentage: 34,
			},
		},
	})
	if err != nil {
		return err
	}

	return enableFlag(flagID)
}

func createGradualRollout30Flag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("gradual_rollout_30"),
		Description: "FLAG:gradual_rollout_30",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	variantKey := "enabled"
	variantID, err := createVariant(flagID, FlagrVariant{
		Key:        variantKey,
		Attachment: map[string]interface{}{"value": true},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "30% rollout",
		RolloutPercent: 30, // Only 30% of users
		Constraints: []FlagrConstraint{
			{
				Property: "country",
				Operator: "EQ",
				Value:    "BR",
			},
		},
	})
	if err != nil {
		return err
	}
	createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: variantKey,
				VariantID:  variantID,
				Percentage: 100,
			},
		},
	})
	if err != nil {
		return err
	}

	return enableFlag(flagID)
}

func createBrazilLaunchFlag() error {
	flagID, err := createFlag(FlagrFlag{
		Key:         randomKey("brazil_launch"),
		Description: "FLAG:brazil_launch",
		Enabled:     true,
	})
	if err != nil {
		return err
	}

	variantKey := "enabled"
	variantID, err := createVariant(flagID, FlagrVariant{
		Key:        variantKey,
		Attachment: map[string]interface{}{"value": true},
	})
	if err != nil {
		return err
	}

	segmentID, err := createSegment(flagID, FlagrSegment{
		Description:    "Brazilian users with valid document",
		RolloutPercent: 100,
		Constraints: []FlagrConstraint{
			{
				Property: "country",
				Operator: "EQ",
				Value:    "BR",
			},
		},
	})
	if err != nil {
		return err
	}

	err = createDistribution(flagID, segmentID, FlagrDistribuitions{
		Distribuitions: []FlagrDistribution{
			{
				VariantKey: variantKey,
				VariantID:  variantID,
				Percentage: 100,
			},
		},
	})
	if err != nil {
		return err
	}

	return enableFlag(flagID)
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func createFlag(flag FlagrFlag) (int64, error) {
	body, _ := json.Marshal(flag)
	resp, err := http.Post(
		flagrEndpoint+"/api/v1/flags",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("createFlag() status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return int64(result["id"].(float64)), nil
}

func createVariant(flagID int64, variant FlagrVariant) (int64, error) {
	body, _ := json.Marshal(variant)
	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/flags/%d/variants", flagrEndpoint, flagID),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("createVariant() status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return int64(result["id"].(float64)), nil
}

func createSegment(flagID int64, segment FlagrSegment) (int64, error) {
	body, _ := json.Marshal(segment)
	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/flags/%d/segments", flagrEndpoint, flagID),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("createSegment() status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return int64(result["id"].(float64)), nil
}

func createDistribution(flagID, segmentID int64, dist FlagrDistribuitions) error {
	body, _ := json.Marshal(dist)
	url := fmt.Sprintf("%s/api/v1/flags/%d/segments/%d/distributions",
		flagrEndpoint, flagID, segmentID)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("createDistribution() status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
