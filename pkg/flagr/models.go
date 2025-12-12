package flagr

import (
	"encoding/json"
	"time"
)

// =======================
// FLAG MODELS (API)
// =======================

// FlagrFlag represents the flag model returned by Flagr API
type FlagrFlag struct {
	ID                 int64          `json:"id"`
	Key                string         `json:"key"`
	Description        string         `json:"description"`
	Enabled            bool           `json:"enabled"`
	Segments           []FlagrSegment `json:"segments"`
	Variants           []FlagrVariant `json:"variants"`
	Tags               []Tag          `json:"tags"`
	DataRecordsEnabled bool           `json:"dataRecordsEnabled"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

// Tag represents a Flagr API tag
type Tag struct {
	Value string `json:"value"`
}

// =======================
// SEGMENTS
// =======================

type FlagrSegment struct {
	ID             int64               `json:"id"`
	Rank           int                 `json:"rank"`
	Description    string              `json:"description"`
	RolloutPercent int64               `json:"rolloutPercent"`
	Constraints    []FlagrConstraint   `json:"constraints"`
	Distributions  []FlagrDistribution `json:"distributions"`
}

// =======================
// CONSTRAINTS
// =======================

type FlagrConstraint struct {
	ID       int64  `json:"id"`
	Property string `json:"property"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// =======================
// DISTRIBUTIONS
// =======================

type FlagrDistribution struct {
	ID        int64 `json:"id"`
	Percent   int64 `json:"percent"`
	VariantID int64 `json:"variantID"`
}

// =======================
// VARIANTS
// =======================

type FlagrVariant struct {
	ID         int64                      `json:"id"`
	Key        string                     `json:"key"`
	Attachment map[string]json.RawMessage `json:"attachment"`
}

// =======================
// EVALUATION REQUEST & RESPONSE
// =======================

type EvaluationRequest struct {
	FlagKey       string                 `json:"flagKey"`
	EntityID      string                 `json:"entityID"`
	EntityType    string                 `json:"entityType"`
	EntityContext map[string]interface{} `json:"entityContext"`
}

// Full evaluation response from Flagr
type EvaluationResponse struct {
	FlagID            int64                      `json:"flagID"`
	FlagKey           string                     `json:"flagKey"`
	SegmentID         int64                      `json:"segmentID"`
	VariantID         int64                      `json:"variantID"`
	VariantKey        string                     `json:"variantKey"`
	VariantAttachment map[string]json.RawMessage `json:"variantAttachment"`
	Timestamp         time.Time                  `json:"timestamp"`
	EvalDebugLog      EvalDebugLog               `json:"evalDebugLog"`
}

// Debug logs used to extract evaluation reason
type EvalDebugLog struct {
	Msg              string            `json:"msg"`
	SegmentDebugLogs []SegmentDebugLog `json:"segmentDebugLogs"`
}

type SegmentDebugLog struct {
	Msg string `json:"msg"`
}

// -----------------------------------------------------------------------------
// Health
// -----------------------------------------------------------------------------

// HealthResponse represents Flagr health-check API response
type HealthResponse struct {
	Status string `json:"status"`
}
