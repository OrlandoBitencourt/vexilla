package vexilla

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/OrlandoBitencourt/vexilla/internal/flagr"
)

// MockFlagrServer is a mock Flagr HTTP server for testing
type MockFlagrServer struct {
	*httptest.Server
	mu    sync.RWMutex
	flags map[int64]domain.Flag
}

// NewMockFlagrServer creates a new mock Flagr server
func NewMockFlagrServer(t *testing.T) *MockFlagrServer {
	mock := &MockFlagrServer{
		flags: make(map[int64]domain.Flag),
	}

	mux := http.NewServeMux()

	// GET /api/v1/flags - List all flags
	mux.HandleFunc("/api/v1/flags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mock.mu.RLock()
		defer mock.mu.RUnlock()

		flags := make([]flagr.FlagrFlag, 0, len(mock.flags))
		for _, flag := range mock.flags {
			flags = append(flags, domainToFlagrFlag(flag))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flags)
	})

	// GET /api/v1/flags/:id - Get single flag
	mux.HandleFunc("/api/v1/flags/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var flagID int64
		if _, err := fmt.Sscanf(r.URL.Path, "/api/v1/flags/%d", &flagID); err != nil {
			http.Error(w, "invalid flag ID", http.StatusBadRequest)
			return
		}

		mock.mu.RLock()
		defer mock.mu.RUnlock()

		flag, ok := mock.flags[flagID]
		if !ok {
			http.Error(w, "flag not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domainToFlagrFlag(flag))
	})

	// POST /api/v1/evaluation - Evaluate flag
	mux.HandleFunc("/api/v1/evaluation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req flagr.EvaluationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		mock.mu.RLock()
		defer mock.mu.RUnlock()

		// Find flag by key
		var flag *domain.Flag
		for _, f := range mock.flags {
			if f.Key == req.FlagKey {
				flag = &f
				break
			}
		}

		if flag == nil {
			http.Error(w, "flag not found", http.StatusNotFound)
			return
		}

		// Simple evaluation - return first variant
		resp := flagr.EvaluationResponse{
			FlagID:            flag.ID,
			FlagKey:           flag.Key,
			VariantKey:        flag.Variants[0].Key,
			VariantID:         flag.Variants[0].ID,
			VariantAttachment: flag.Variants[0].Attachment,
		}

		if len(flag.Segments) > 0 {
			resp.SegmentID = flag.Segments[0].ID
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/health - Health check
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
	})

	mock.Server = httptest.NewServer(mux)
	return mock
}

// AddFlag adds a flag to the mock server
func (m *MockFlagrServer) AddFlag(flag domain.Flag) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flags[flag.ID] = flag
}

// RemoveFlag removes a flag from the mock server
func (m *MockFlagrServer) RemoveFlag(flagID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.flags, flagID)
}

// domainToFlagrFlag converts domain.Flag to flagr.FlagrFlag
func domainToFlagrFlag(f domain.Flag) flagr.FlagrFlag {
	segments := make([]flagr.FlagrSegment, len(f.Segments))
	for i, s := range f.Segments {
		constraints := make([]flagr.FlagrConstraint, len(s.Constraints))
		for j, c := range s.Constraints {
			constraints[j] = flagr.FlagrConstraint{
				ID:       c.ID,
				Property: c.Property,
				Operator: string(c.Operator),
				Value:    fmt.Sprintf("%v", c.Value),
			}
		}

		distributions := make([]flagr.FlagrDistribution, len(s.Distributions))
		for j, d := range s.Distributions {
			distributions[j] = flagr.FlagrDistribution{
				ID:        d.ID,
				Percent:   int64(d.Percent),
				VariantID: d.VariantID,
			}
		}

		segments[i] = flagr.FlagrSegment{
			ID:             s.ID,
			Rank:           s.Rank,
			Description:    s.Description,
			RolloutPercent: int64(s.RolloutPercent),
			Constraints:    constraints,
			Distributions:  distributions,
		}
	}

	variants := make([]flagr.FlagrVariant, len(f.Variants))
	for i, v := range f.Variants {
		variants[i] = flagr.FlagrVariant{
			ID:         v.ID,
			Key:        v.Key,
			Attachment: v.Attachment,
		}
	}

	tags := make([]flagr.Tag, len(f.Tags))
	for i, t := range f.Tags {
		tags[i] = flagr.Tag{Value: t.Value}
	}

	return flagr.FlagrFlag{
		ID:          f.ID,
		Key:         f.Key,
		Description: f.Description,
		Enabled:     f.Enabled,
		Segments:    segments,
		Variants:    variants,
		Tags:        tags,
		UpdatedAt:   f.UpdatedAt,
	}
}
