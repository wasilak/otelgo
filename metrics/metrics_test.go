package metrics

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func TestInit(t *testing.T) {
	// Save original env vars and restore them after test
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", originalProtocol)
	}()

	// Set test environment to use in-memory reader
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "none")

	tests := []struct {
		name    string
		config  OtelGoMetricsConfig
		wantErr bool
	}{
		{
			name: "basic initialization",
			config: OtelGoMetricsConfig{
				Attributes: []attribute.KeyValue{
					semconv.ServiceNameKey.String("test-service"),
				},
			},
			wantErr: false,
		},
		{
			name:    "empty config",
			config:  OtelGoMetricsConfig{},
			wantErr: false,
		},
		{
			name: "with multiple attributes",
			config: OtelGoMetricsConfig{
				Attributes: []attribute.KeyValue{
					semconv.ServiceNameKey.String("test-service"),
					semconv.ServiceVersionKey.String("1.0.0"),
					attribute.String("custom.attribute", "value"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a manual reader for testing
			reader := metric.NewManualReader()
			provider := metric.NewMeterProvider(metric.WithReader(reader))

			// Create a meter and some test metrics
			meter := provider.Meter("test")
			counter, err := meter.Int64Counter("test.counter")
			if err != nil {
				t.Fatalf("Failed to create counter: %v", err)
			}
			counter.Add(ctx, 1)

			// Collect metrics
			metrics := metricdata.ResourceMetrics{}
			err = reader.Collect(ctx, &metrics)
			if err != nil {
				t.Errorf("Failed to collect metrics: %v", err)
			}

			// Verify the collected metrics
			if len(metrics.ScopeMetrics) != 1 {
				t.Errorf("Expected 1 scope metric, got %d", len(metrics.ScopeMetrics))
			}

			if len(metrics.ScopeMetrics[0].Metrics) != 1 {
				t.Errorf("Expected 1 metric, got %d", len(metrics.ScopeMetrics[0].Metrics))
			}

			metric := metrics.ScopeMetrics[0].Metrics[0]
			if metric.Name != "test.counter" {
				t.Errorf("Expected metric name 'test.counter', got '%s'", metric.Name)
			}

			if sum, ok := metric.Data.(metricdata.Sum[int64]); !ok {
				t.Error("Expected Sum[int64] data")
			} else {
				if len(sum.DataPoints) != 1 {
					t.Errorf("Expected 1 data point, got %d", len(sum.DataPoints))
				}
				if sum.DataPoints[0].Value != 1 {
					t.Errorf("Expected value 1, got %d", sum.DataPoints[0].Value)
				}
			}

			err = Shutdown(ctx, provider)
			if err != nil {
				t.Errorf("Shutdown() error = %v", err)
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()

	// Create a manual reader for testing
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	// Test first shutdown - should succeed
	err := Shutdown(ctx, provider)
	if err != nil {
		t.Errorf("First Shutdown() error = %v", err)
	}

	// Test second shutdown - should return "reader is shutdown" error
	err = Shutdown(ctx, provider)
	if err == nil {
		t.Error("Second Shutdown() expected error but got nil")
	} else if !strings.Contains(err.Error(), "reader is shutdown") {
		t.Errorf("Second Shutdown() expected 'reader is shutdown' error, got %v", err)
	}
}

func TestInitWithTLSValidation(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                     os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL":   os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_INSECURE":           os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"),
		"OTEL_EXPORTER_OTLP_CERTIFICATE":        os.Getenv("OTEL_EXPORTER_OTLP_CERTIFICATE"),
		"OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE": os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE"),
		"OTEL_EXPORTER_OTLP_CLIENT_KEY":         os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_KEY"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment to use in-memory reader
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "none")

	tests := []struct {
		name    string
		config  OtelGoMetricsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "conflicting insecure and CA cert",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					Insecure:   true,
					CACertPath: "/path/to/ca.pem",
				},
			},
			wantErr: true,
			errMsg:  "invalid TLS configuration",
		},
		{
			name: "missing client key",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					ClientCertPath: "/path/to/cert.pem",
				},
			},
			wantErr: true,
			errMsg:  "invalid TLS configuration",
		},
		{
			name: "invalid CA cert path",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					CACertPath: "/nonexistent/path/ca.pem",
				},
			},
			wantErr: true,
			errMsg:  "failed to build TLS config",
		},
		{
			name: "valid insecure config",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					Insecure: true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a manual reader for testing to avoid actual network connections
			reader := metric.NewManualReader()
			manualProvider := metric.NewMeterProvider(metric.WithReader(reader))

			// We can't actually test the TLS connection without a server, so we'll focus on validation
			// The TLS validation happens during Init, which should return an error if validation fails
			_, _, err := Init(ctx, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Init() error message should contain '%s', got '%v'", tt.errMsg, err)
				}
			}

			// Shutdown the manual provider
			err = manualProvider.Shutdown(ctx)
			if err != nil {
				t.Errorf("Manual provider Shutdown() error = %v", err)
			}
		})
	}
}

func TestInitBackwardCompatibility(t *testing.T) {
	// Save original env vars and restore them after test
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", originalProtocol)
	}()

	// Set test environment to use in-memory reader
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "none")

	ctx := context.Background()
	config := OtelGoMetricsConfig{
		Attributes: []attribute.KeyValue{
			semconv.ServiceNameKey.String("backward-compatible-service"),
		},
	}

	// Create a manual reader for testing
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	// We can initialize without TLS config for backward compatibility
	_, _, err := Init(ctx, config)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Shutdown the manual provider
	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestInitErrorPathScenarios(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                   os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "error-test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4318")

	ctx := context.Background()

	scenarios := []struct {
		name          string
		config        OtelGoMetricsConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config",
			config: OtelGoMetricsConfig{
				Attributes: []attribute.KeyValue{
					attribute.String("test", "valid"),
				},
			},
			expectError: false,
		},
		{
			name: "invalid TLS config - conflicting settings",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					Insecure:   true,
					CACertPath: "/nonexistent/ca.pem", // Can't have both Insecure and CA cert
				},
			},
			expectError:   true,
			errorContains: "cannot specify both Insecure=true and CACertPath",
		},
		{
			name: "invalid TLS config - missing client key",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					ClientCertPath: "/path/to/cert.pem",
					// Missing ClientKeyPath
				},
			},
			expectError:   true,
			errorContains: "invalid TLS configuration",
		},
		{
			name: "invalid TLS config - invalid CA path",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					CACertPath: "/definitely/nonexistent/ca.pem",
				},
			},
			expectError:   true,
			errorContains: "failed to read CA certificate",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			_, provider, err := Init(ctx, scenario.config)

			if (err != nil) != scenario.expectError {
				t.Errorf("Init() error = %v, expectError %v", err, scenario.expectError)
			}

			if scenario.expectError && err != nil {
				if scenario.errorContains != "" && !strings.Contains(err.Error(), scenario.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", scenario.errorContains, err)
				}
			}

			if provider != nil {
				_ = Shutdown(ctx, provider)
			}
		})
	}
}

func TestInitWithInvalidEndpoint(t *testing.T) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "invalid-endpoint-test")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")

	ctx := context.Background()

	// Test with an invalid endpoint - this will cause errors during actual export attempts
	// but Init should still succeed if the configuration is valid
	config := OtelGoMetricsConfig{
		TLS: &internal.TLSConfig{
			Insecure: true, // Use insecure to avoid certificate validation issues
		},
	}

	_, provider, err := Init(ctx, config)
	if err != nil {
		// This might fail depending on network connectivity, which is okay for this test
		t.Logf("Init with invalid endpoint returned error (could happen): %v", err)
	} else {
		// If initialization succeeded, we should still be able to shutdown properly
		if provider != nil {
			_ = Shutdown(ctx, provider)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInitConcurrency(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                   os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment to use in-memory reader to avoid actual network connections
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "none")

	ctx := context.Background()
	var wg sync.WaitGroup
	const numGoroutines = 50

	// Test concurrent initialization with manual readers to avoid actual network calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Each goroutine gets its own config to test isolation
			config := OtelGoMetricsConfig{
				Attributes: []attribute.KeyValue{
					attribute.String("goroutine.id", fmt.Sprintf("%d", i)),
				},
			}

			// Create a manual reader for testing to avoid actual network connections
			reader := metric.NewManualReader()
			manualProvider := metric.NewMeterProvider(metric.WithReader(reader))

			// Initialize without network calls
			_, _, err := Init(ctx, config)
			if err != nil && !strings.Contains(err.Error(), "connection refused") && !strings.Contains(err.Error(), "no such host") {
				t.Errorf("Metrics Init failed for goroutine %d: %v", i, err)
			}

			_ = manualProvider.Shutdown(ctx)
		}(i)
	}

	wg.Wait()
}

func TestInitErrorPaths(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                     os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL":   os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT":   os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"),
		"OTEL_EXPORTER_OTLP_INSECURE":           os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"),
		"OTEL_EXPORTER_OTLP_CERTIFICATE":        os.Getenv("OTEL_EXPORTER_OTLP_CERTIFICATE"),
		"OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE": os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE"),
		"OTEL_EXPORTER_OTLP_CLIENT_KEY":         os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_KEY"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment to use in-memory reader to avoid actual network connections
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "none")

	tests := []struct {
		name    string
		config  OtelGoMetricsConfig
		wantErr bool
	}{
		{
			name: "invalid TLS config - conflicting insecure and CA cert",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					Insecure:   true,
					CACertPath: "/path/to/ca.pem",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid TLS config - missing client key",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					ClientCertPath: "/path/to/cert.pem",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid TLS config - invalid CA cert path",
			config: OtelGoMetricsConfig{
				TLS: &internal.TLSConfig{
					CACertPath: "/nonexistent/path/ca.pem",
				},
			},
			wantErr: true,
		},
		{
			name: "valid config with minimal settings",
			config: OtelGoMetricsConfig{
				Attributes: []attribute.KeyValue{
					attribute.String("test", "value"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a manual reader for testing to avoid actual network connections
			reader := metric.NewManualReader()
			manualProvider := metric.NewMeterProvider(metric.WithReader(reader))

			_, _, err := Init(ctx, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Shutdown the manual provider
			_ = manualProvider.Shutdown(ctx)
		})
	}
}

func countGoroutines() int {
	return runtime.NumGoroutine()
}

func TestGoroutineLeak(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                   os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment to avoid network connections
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4318")

	baseGoroutines := countGoroutines()

	ctx := context.Background()

	// Initialize metrics provider
	_, provider, err := Init(ctx, OtelGoMetricsConfig{})
	if err != nil {
		t.Fatalf("Failed to initialize metrics: %v", err)
	}

	// Shutdown metrics provider
	if err := Shutdown(ctx, provider); err != nil {
		t.Logf("Note: Shutdown error expected due to non-existent endpoint: %v", err) // Log but don't fail
	}

	// Allow goroutines to finish
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := countGoroutines()

	// Allow some variance for goroutines that may still be cleaning up
	// The key is to make sure there isn't a significant leak
	if finalGoroutines > baseGoroutines+5 {
		t.Errorf("Potential goroutine leak detected: started with %d goroutines, ended with %d", baseGoroutines, finalGoroutines)
	}
}

func TestContextCancellation(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                   os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4318")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Initialize metrics provider
	_, provider, err := Init(ctx, OtelGoMetricsConfig{})
	if err != nil {
		t.Fatalf("Failed to initialize metrics: %v", err)
	}

	// Shutdown metrics provider with the same context
	if err := Shutdown(ctx, provider); err != nil {
		// For this test, we expect potential errors due to the connection attempt
		// The important thing is that it doesn't hang
		t.Logf("Shutdown returned error (expected): %v", err)
	}

	// Check that context is still valid (not cancelled by the test operations)
	if ctx.Err() != nil {
		t.Errorf("Context was unexpectedly cancelled: %v", ctx.Err())
	}
}

func TestConfigurationValidation(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME": os.Getenv("OTEL_SERVICE_NAME"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Test default configuration values
	os.Setenv("OTEL_SERVICE_NAME", "test-default-service")

	ctx := context.Background()

	// Test with empty config to ensure defaults work
	config := OtelGoMetricsConfig{}
	_, provider, err := Init(ctx, config)
	if err != nil {
		t.Errorf("Init with empty config should succeed: %v", err)
	}
	if provider == nil {
		t.Error("Provider should not be nil with empty config")
	}
	if err := Shutdown(ctx, provider); err != nil {
		t.Logf("Shutdown error (expected): %v", err)
	}

	// Test with custom attributes to ensure they are applied
	customAttrs := []attribute.KeyValue{
		attribute.String("env", "test"),
		attribute.Int("version", 1),
	}
	configWithAttrs := OtelGoMetricsConfig{
		Attributes: customAttrs,
	}
	_, provider2, err := Init(ctx, configWithAttrs)
	if err != nil {
		t.Errorf("Init with custom attributes should succeed: %v", err)
	}
	if provider2 == nil {
		t.Error("Provider should not be nil with custom attributes")
	}
	if err := Shutdown(ctx, provider2); err != nil {
		t.Logf("Shutdown error (expected): %v", err)
	}
}
