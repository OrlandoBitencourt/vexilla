package flagr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
)

// HTTPClient implements Client interface using HTTP
type HTTPClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
	maxRetries int
}

// NewHTTPClient creates a new Flagr HTTP client
func NewHTTPClient(config Config) *HTTPClient {
	return &HTTPClient{
		endpoint: config.Endpoint,
		apiKey:   config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		maxRetries: config.MaxRetries,
	}
}

// GetAllFlags fetches all flags from Flagr
func (c *HTTPClient) GetAllFlags(ctx context.Context) ([]domain.Flag, error) {
	url := fmt.Sprintf("%s/api/v1/flags", c.endpoint)

	var flagrFlags []FlagrFlag
	if err := c.doRequest(ctx, "GET", url, nil, &flagrFlags); err != nil {
		return nil, fmt.Errorf("failed to fetch flags: %w", err)
	}

	// Fetch detailed info for each flag (includes segments, constraints, etc.)
	detailedFlags := []domain.Flag{}
	for _, flag := range flagrFlags {
		detailedFlag, err := c.GetFlag(ctx, flag.ID)
		if err != nil {
			// Log but continue with partial data
			fmt.Printf("warning: failed to fetch details for flag %d: %v\n", flag.ID, err)
			continue
		}
		detailedFlags = append(detailedFlags, *detailedFlag)
	}

	return detailedFlags, nil
}

// GetFlag fetches a single flag by ID with full details
func (c *HTTPClient) GetFlag(ctx context.Context, flagID int64) (*domain.Flag, error) {
	url := fmt.Sprintf("%s/api/v1/flags/%d", c.endpoint, flagID)

	// resposta do flagr (raw)
	var apiFlag FlagrFlag
	if err := c.doRequest(ctx, "GET", url, nil, &apiFlag); err != nil {
		return nil, fmt.Errorf("failed to fetch flag %d: %w", flagID, err)
	}

	// converter para domain.Flag usando o adapter completo
	domainFlag := FlagToDomain(&apiFlag)
	return &domainFlag, nil
}

// EvaluateFlag evaluates a flag remotely using Flagr
func (c *HTTPClient) EvaluateFlag(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
	url := fmt.Sprintf("%s/api/v1/evaluation", c.endpoint)

	req := EvaluationContextFromDomain(flagKey, evalCtx)

	var resp EvaluationResponse
	if err := c.doRequest(ctx, "POST", url, req, &resp); err != nil {
		return nil, domain.NewEvaluationError(flagKey, "remote evaluation failed", err)
	}

	result := EvaluationResultToDomain(resp)
	return &result, nil
}

// HealthCheck verifies Flagr is reachable
func (c *HTTPClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/health", c.endpoint)

	var health HealthResponse
	if err := c.doRequest(ctx, "GET", url, nil, &health); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if health.Status != "OK" {
		return fmt.Errorf("unhealthy status: %s", health.Status)
	}

	return nil
}

// doRequest performs HTTP request with retries
func (c *HTTPClient) doRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * 500 * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := c.doSingleRequest(ctx, method, url, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on certain errors
		if !c.shouldRetry(err) {
			return lastErr
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doSingleRequest performs a single HTTP request
func (c *HTTPClient) doSingleRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w (body: %s)", err, string(respBody))
		}
	}

	return nil
}

// shouldRetry determines if request should be retried
func (c *HTTPClient) shouldRetry(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		// Retry on 5xx and 429 (rate limit)
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == 429
	}
	// Retry on network errors
	return true
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}
