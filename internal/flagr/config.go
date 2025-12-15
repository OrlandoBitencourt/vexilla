package flagr

import "time"

// Config defines settings for the Flagr HTTP client
type Config struct {
	Endpoint   string        // Base URL of Flagr (ex: http://localhost:18000)
	APIKey     string        // Optional: Authorization header as Bearer <APIKey>
	Timeout    time.Duration // HTTP timeout for requests
	MaxRetries int           // Number of retry attempts for failed HTTP requests
}
