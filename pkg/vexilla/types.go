package vexilla

import "time"

// Flag represents a feature flag from Flagr
type Flag struct {
	ID          int64     `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	Segments    []Segment `json:"segments"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`

	// Vexilla metadata
	LastRefreshed time.Time `json:"-"`
	CacheExpiry   time.Time `json:"-"`
}

// Segment represents a flag segment with constraints and distributions
type Segment struct {
	ID             int64          `json:"id"`
	Description    string         `json:"description,omitempty"`
	RolloutPercent int            `json:"rolloutPercent"` // 0-100
	Rank           int            `json:"rank"`
	Constraints    []Constraint   `json:"constraints"`
	Distributions  []Distribution `json:"distributions"`
}

// Constraint represents a segment constraint
type Constraint struct {
	ID       int64       `json:"id"`
	Property string      `json:"property"`
	Operator string      `json:"operator"` // EQ, NEQ, IN, NOTIN, MATCHES, etc.
	Value    interface{} `json:"value"`
}

// Distribution represents variant distribution
type Distribution struct {
	ID                int64       `json:"id"`
	Percent           int         `json:"percent"` // 0-100
	VariantID         int64       `json:"variantID"`
	VariantKey        string      `json:"variantKey"`
	VariantAttachment interface{} `json:"variantAttachment,omitempty"`
}

// EvaluationContext holds user/request context for flag evaluation
type EvaluationContext struct {
	EntityID   string                 `json:"entityID"`
	EntityType string                 `json:"entityType,omitempty"`
	Context    map[string]interface{} `json:"entityContext,omitempty"`
}

// EvaluationResult represents the result of a flag evaluation
type EvaluationResult struct {
	FlagID            int64       `json:"flagID"`
	FlagKey           string      `json:"flagKey"`
	SegmentID         int64       `json:"segmentID,omitempty"`
	VariantID         int64       `json:"variantID"`
	VariantKey        string      `json:"variantKey"`
	VariantAttachment interface{} `json:"variantAttachment,omitempty"`

	// Vexilla metadata
	EvaluatedLocally bool          `json:"-"`
	EvaluationTime   time.Duration `json:"-"`
}

// Stats represents cache statistics
type Stats struct {
	KeysAdded        uint64    `json:"keys_added"`
	KeysUpdated      uint64    `json:"keys_updated"`
	KeysEvicted      uint64    `json:"keys_evicted"`
	CostAdded        uint64    `json:"cost_added"`
	CostEvicted      uint64    `json:"cost_evicted"`
	SetsDropped      uint64    `json:"sets_dropped"`
	SetsRejected     uint64    `json:"sets_rejected"`
	GetsKept         uint64    `json:"gets_kept"`
	GetsDropped      uint64    `json:"gets_dropped"`
	HitRatio         float64   `json:"hit_ratio"`
	LastRefresh      time.Time `json:"last_refresh"`
	ConsecutiveFails int       `json:"consecutive_fails"`
	CircuitOpen      bool      `json:"circuit_open"`
}
