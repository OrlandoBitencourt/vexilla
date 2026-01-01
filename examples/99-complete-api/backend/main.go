package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Server holds the application state
type Server struct {
	vexilla     *vexilla.Client
	rateLimiter *RateLimiter
	tracer      trace.Tracer
}

// RateLimiter simple in-memory rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this key
	times := rl.requests[key]

	// Filter out old requests
	validTimes := []time.Time{}
	for _, t := range times {
		if t.After(cutoff) {
			validTimes = append(validTimes, t)
		}
	}

	// Check if limit exceeded
	if len(validTimes) >= rl.limit {
		return false
	}

	// Add current request
	validTimes = append(validTimes, now)
	rl.requests[key] = validTimes

	return true
}

// BuildEvalContext creates a Vexilla context from request headers
func BuildEvalContext(c *gin.Context) vexilla.Context {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	return vexilla.NewContext(userID).
		WithAttribute("role", c.GetHeader("X-User-Role")).
		WithAttribute("country", c.GetHeader("X-Country")).
		WithAttribute("cpf", c.GetHeader("X-CPF"))
}

// DeterministicBucketFromCPF calculates a bucket (0-99) from CPF
func DeterministicBucketFromCPF(cpf string) (int, error) {
	// Remove non-digits
	re := regexp.MustCompile(`[^\d]`)
	cleanCPF := re.ReplaceAllString(cpf, "")

	if len(cleanCPF) != 11 {
		return 0, fmt.Errorf("CPF inválido: deve ter 11 dígitos")
	}

	// Hash the CPF to get consistent bucket
	hash := sha256.Sum256([]byte(cleanCPF))

	// Use first 8 bytes to create a uint64
	bucket64 := binary.BigEndian.Uint64(hash[:8])

	// Modulo 100 to get 0-99
	bucket := int(bucket64 % 100)

	return bucket, nil
}

// initJaeger initializes OpenTelemetry with Jaeger exporter
func initJaeger(ctx context.Context) (func(context.Context) error, error) {
	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("vexilla-demo-api"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func main() {
	fmt.Println("🏴 Vexilla Demo API")
	fmt.Println("===================")
	fmt.Println()

	ctx := context.Background()

	// Initialize Jaeger tracing
	fmt.Println("🔍 Initializing Jaeger tracing...")
	shutdown, err := initJaeger(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Jaeger: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("⚠️  Failed to shutdown tracer: %v", err)
		}
	}()
	fmt.Println("✅ Jaeger tracing initialized (http://localhost:16686)")

	// Create Vexilla client (will automatically use the global OpenTelemetry tracer)
	fmt.Println("📦 Creating Vexilla client with OpenTelemetry...")
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(30*time.Second),
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create Vexilla client: %v", err)
	}

	// Start client
	fmt.Println("🚀 Starting Vexilla client...")
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start Vexilla client: %v", err)
	}
	defer client.Stop()

	// Initialize server
	tracer := otel.Tracer("vexilla-demo-api")
	server := &Server{
		vexilla:     client,
		rateLimiter: NewRateLimiter(10, time.Minute), // 10 requests per minute
		tracer:      tracer,
	}

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-User-ID", "X-CPF", "X-User-Role", "X-Country"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Custom logging middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		log.Printf("[%d] %s %s - %v\n", statusCode, c.Request.Method, path, latency)
	})

	// OpenTelemetry tracing middleware
	r.Use(server.TracingMiddleware())

	// Apply rate limit middleware globally
	r.Use(server.RateLimitMiddleware())

	// Routes
	r.GET("/health", server.healthHandler)
	r.POST("/checkout", server.checkoutHandler)
	r.GET("/flags/snapshot", server.flagsSnapshotHandler)

	// Admin routes
	admin := r.Group("/admin")
	{
		admin.POST("/flags/invalidate-all", server.invalidateAllFlagsHandler)
		admin.POST("/flags/:flagKey", server.invalidateFlagHandler)
		admin.GET("/flags/metrics", server.flagsMetricsHandler)
	}

	// Start server
	fmt.Println("✅ Server ready!")
	fmt.Println("🌐 Listening on http://localhost:9080")
	fmt.Println()
	fmt.Println("📋 Available routes:")
	fmt.Println("   GET  /health")
	fmt.Println("   POST /checkout")
	fmt.Println("   GET  /flags/snapshot")
	fmt.Println("   POST /admin/flags/invalidate-all")
	fmt.Println("   POST /admin/flags/:flagKey")
	fmt.Println("   GET  /admin/flags/metrics")
	fmt.Println()

	if err := r.Run(":9080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// Middleware
func (s *Server) TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		if c.FullPath() == "" {
			spanName = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		}

		ctx, span := s.tracer.Start(ctx, spanName)
		defer span.End()

		// Add request attributes to span
		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			semconv.HTTPTargetKey.String(c.Request.URL.Path),
			semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
		)

		// Add custom attributes from headers
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			span.SetAttributes(semconv.EnduserIDKey.String(userID))
		}
		if role := c.GetHeader("X-User-Role"); role != "" {
			span.SetAttributes(semconv.EnduserRoleKey.String(role))
		}

		// Update context in request
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Add response status to span
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(c.Writer.Status()))
	}
}

func (s *Server) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		evalCtx := BuildEvalContext(c)

		// Check if rate limit is enabled via flag
		enabled := s.vexilla.Bool(context.Background(), "api.rate_limit.enabled", evalCtx)

		if enabled {
			clientIP := c.ClientIP()
			if !s.rateLimiter.IsAllowed(clientIP) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "Too Many Requests",
					"message": "Rate limit exceeded. Try again later.",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// Handlers
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "vexilla-demo-api",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) checkoutHandler(c *gin.Context) {
	cpf := c.GetHeader("X-CPF")
	if cpf == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "CPF obrigatório (header X-CPF)",
		})
		return
	}

	// Calculate deterministic bucket from CPF
	bucket, err := DeterministicBucketFromCPF(cpf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "CPF inválido",
			"details": err.Error(),
		})
		return
	}

	evalCtx := BuildEvalContext(c)

	// Check kill switch first
	v2Enabled := s.vexilla.Bool(context.Background(), "api.checkout.v2", evalCtx)
	if !v2Enabled {
		s.checkoutV1(c, bucket)
		return
	}

	// Check rollout percentage
	result, err := s.vexilla.Evaluate(context.Background(), "api.checkout.rollout", evalCtx)
	if err != nil {
		log.Printf("Error evaluating rollout flag: %v\n", err)
		s.checkoutV1(c, bucket)
		return
	}

	// Get rollout percentage from variant attachment
	rollout := 0
	if result != nil && result.VariantAttachment != nil {
		if valueData, ok := result.VariantAttachment["value"]; ok {
			var val float64
			if err := json.Unmarshal(valueData, &val); err == nil {
				rollout = int(val)
			}
		}
	}

	// Determine version based on bucket and rollout
	version := "v1"
	if bucket < rollout {
		version = "v2"
	}

	if version == "v2" {
		s.checkoutV2(c, bucket)
	} else {
		s.checkoutV1(c, bucket)
	}
}

func (s *Server) checkoutV1(c *gin.Context, bucket int) {
	c.JSON(http.StatusOK, gin.H{
		"version": "v1",
		"message": "Checkout Legacy",
		"details": gin.H{
			"cpf":    c.GetHeader("X-CPF"),
			"bucket": bucket,
			"ui":     "classic",
			"color":  "#3B82F6",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) checkoutV2(c *gin.Context, bucket int) {
	c.JSON(http.StatusOK, gin.H{
		"version": "v2",
		"message": "Checkout Modernizado",
		"details": gin.H{
			"cpf":    c.GetHeader("X-CPF"),
			"bucket": bucket,
			"ui":     "modern",
			"color":  "#10B981",
		},
		"features": []string{
			"One-click checkout",
			"Apple Pay / Google Pay",
			"Saved payment methods",
			"Express shipping",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) flagsSnapshotHandler(c *gin.Context) {
	evalCtx := BuildEvalContext(c)

	// Get all flags
	flags := map[string]interface{}{
		"api.checkout.v2":        s.vexilla.Bool(context.Background(), "api.checkout.v2", evalCtx),
		"api.rate_limit.enabled": s.vexilla.Bool(context.Background(), "api.rate_limit.enabled", evalCtx),
		"api.kill_switch":        s.vexilla.Bool(context.Background(), "api.kill_switch", evalCtx),
		"frontend.new_ui":        s.vexilla.Bool(context.Background(), "frontend.new_ui", evalCtx),
		"frontend.beta_banner":   s.vexilla.Bool(context.Background(), "frontend.beta_banner", evalCtx),
	}

	// Get rollout value
	rolloutResult, _ := s.vexilla.Evaluate(context.Background(), "api.checkout.rollout", evalCtx)
	rolloutValue := 0
	if rolloutResult != nil && rolloutResult.VariantAttachment != nil {
		if valueData, ok := rolloutResult.VariantAttachment["value"]; ok {
			var val float64
			if err := json.Unmarshal(valueData, &val); err == nil {
				rolloutValue = int(val)
			}
		}
	}
	flags["api.checkout.rollout"] = rolloutValue

	// Get bucket if CPF provided
	var bucket *int
	if cpf := c.GetHeader("X-CPF"); cpf != "" {
		if b, err := DeterministicBucketFromCPF(cpf); err == nil {
			bucket = &b
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"flags": flags,
		"context": gin.H{
			"user_id": c.GetHeader("X-User-ID"),
			"cpf":     c.GetHeader("X-CPF"),
			"role":    c.GetHeader("X-User-Role"),
			"country": c.GetHeader("X-Country"),
			"bucket":  bucket,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) invalidateAllFlagsHandler(c *gin.Context) {
	ctx := context.Background()
	if err := s.vexilla.InvalidateAll(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to invalidate all flags",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "All flags cache invalidated successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) invalidateFlagHandler(c *gin.Context) {
	flagKey := c.Param("flagKey")
	ctx := context.Background()

	if err := s.vexilla.InvalidateFlag(ctx, flagKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to invalidate flag",
			"flag":    flagKey,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Flag '%s' invalidated successfully", flagKey),
		"flag":      flagKey,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) flagsMetricsHandler(c *gin.Context) {
	// Simple metrics - in real app this would come from Vexilla
	c.JSON(http.StatusOK, gin.H{
		"cache_hits":      120,
		"cache_miss":      8,
		"flags_loaded":    6,
		"circuit_breaker": "closed",
		"last_refresh":    time.Now().Add(-15 * time.Second).Format(time.RFC3339),
		"uptime_seconds":  time.Now().Unix(),
	})
}
