package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FlagrClient handles HTTP communication with Flagr server
type FlagrClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
	tracer     trace.Tracer
}

// NewFlagrClient creates a new Flagr HTTP client
func NewFlagrClient(endpoint string, apiKey string, timeout time.Duration) *FlagrClient {
	return &FlagrClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		tracer: otel.Tracer("vexilla.client"),
	}
}

// GetFlags fetches all flags from Flagr
func (c *FlagrClient) GetFlags(ctx context.Context) ([]vexilla.Flag, error) {
	ctx, span := c.tracer.Start(ctx, "flagr.get_flags")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/flags", c.endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch flags: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("flagr returned status %d: %s", resp.StatusCode, string(body))
	}

	var flags []vexilla.Flag
	if err := json.NewDecoder(resp.Body).Decode(&flags); err != nil {
		return nil, fmt.Errorf("failed to decode flags: %w", err)
	}

	span.SetAttributes(attribute.Int("flags.count", len(flags)))
	return flags, nil
}

// PostEvaluation evaluates a flag via Flagr API
func (c *FlagrClient) PostEvaluation(ctx context.Context, flagKey string, evalCtx vexilla.EvaluationContext) (*vexilla.EvaluationResult, error) {
	ctx, span := c.tracer.Start(ctx, "flagr.post_evaluation",
		trace.WithAttributes(
			attribute.String("flag.key", flagKey),
			attribute.String("entity.id", evalCtx.EntityID),
		),
	)
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/evaluation", c.endpoint)

	reqBody := map[string]interface{}{
		"flagKey":       flagKey,
		"entityID":      evalCtx.EntityID,
		"entityType":    evalCtx.EntityType,
		"entityContext": evalCtx.Context,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to evaluate flag: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("flagr returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		FlagID            int64       `json:"flagID"`
		FlagKey           string      `json:"flagKey"`
		SegmentID         int64       `json:"segmentID"`
		VariantID         int64       `json:"variantID"`
		VariantKey        string      `json:"variantKey"`
		VariantAttachment interface{} `json:"variantAttachment"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode evaluation result: %w", err)
	}

	span.SetAttributes(
		attribute.String("variant.key", result.VariantKey),
		attribute.Int64("variant.id", result.VariantID),
	)

	return &vexilla.EvaluationResult{
		FlagID:            result.FlagID,
		FlagKey:           result.FlagKey,
		SegmentID:         result.SegmentID,
		VariantID:         result.VariantID,
		VariantKey:        result.VariantKey,
		VariantAttachment: result.VariantAttachment,
		EvaluatedLocally:  false,
	}, nil
}

// Health checks Flagr server health
func (c *FlagrClient) Health(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "flagr.health")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/health", c.endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("flagr health check failed with status %d", resp.StatusCode)
	}

	return nil
}
