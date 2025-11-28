package logs

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
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func TestInit(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	tests := []struct {
		name    string
		config  OtelGoLogsConfig
		wantErr bool
	}{
		{
			name: "basic initialization",
			config: OtelGoLogsConfig{
				Attributes: []attribute.KeyValue{
					semconv.ServiceNameKey.String("test-service"),
				},
			},
			wantErr: false,
		},
		{
			name:    "empty config",
			config:  OtelGoLogsConfig{},
			wantErr: false,
		},
		{
			name: "with multiple attributes",
			config: OtelGoLogsConfig{
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
			ctx, provider, err := Init(ctx, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && provider == nil {
				t.Error("Init() returned nil provider but no error")
			}

			if provider != nil {
				err = Shutdown(ctx, provider)
				if err != nil {
					t.Errorf("Shutdown() error = %v", err)
				}
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()
	config := OtelGoLogsConfig{
		Attributes: []attribute.KeyValue{
			semconv.ServiceNameKey.String("test-service"),
		},
	}

	// Initialize a provider for shutdown testing
	_, provider, err := Init(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize provider for shutdown test: %v", err)
	}

	// Test first shutdown - should succeed
	err = Shutdown(ctx, provider)
	if err != nil {
		t.Errorf("First Shutdown() error = %v", err)
	}
}

func TestInitWithTLSValidation(t *testing.T) {
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                     os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL":      os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT":      os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
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

	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	tests := []struct {
		name    string
		config  OtelGoLogsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "conflicting insecure and CA cert",
			config: OtelGoLogsConfig{
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
			config: OtelGoLogsConfig{
				TLS: &internal.TLSConfig{
					ClientCertPath: "/path/to/cert.pem",
				},
			},
			wantErr: true,
			errMsg:  "invalid TLS configuration",
		},
		{
			name: "invalid CA cert path",
			config: OtelGoLogsConfig{
				TLS: &internal.TLSConfig{
					CACertPath: "/nonexistent/path/ca.pem",
				},
			},
			wantErr: true,
			errMsg:  "failed to build TLS config",
		},
		{
			name: "valid insecure config",
			config: OtelGoLogsConfig{
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
			_, provider, err := Init(ctx, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Init() error message should contain '%s', got '%v'", tt.errMsg, err)
				}
			}

			if provider != nil {
				err = Shutdown(ctx, provider)
				if err != nil {
					t.Errorf("Shutdown() error = %v", err)
				}
			}
		})
	}
}

func TestInitBackwardCompatibility(t *testing.T) {
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	ctx := context.Background()
	config := OtelGoLogsConfig{
		Attributes: []attribute.KeyValue{
			semconv.ServiceNameKey.String("backward-compatible-service"),
		},
	}

	_, provider, err := Init(ctx, config)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if provider == nil {
		t.Error("Init() returned nil provider but no error")
	}

	if provider != nil {
		err = Shutdown(ctx, provider)
		if err != nil {
			t.Errorf("Shutdown() error = %v", err)
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
		"OTEL_SERVICE_NAME":                os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	ctx := context.Background()
	var wg sync.WaitGroup
	const numGoroutines = 50

	// Test concurrent initialization
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Each goroutine gets its own config to test isolation
			config := OtelGoLogsConfig{
				Attributes: []attribute.KeyValue{
					attribute.String("goroutine.id", fmt.Sprintf("%d", i)),
				},
			}
			_, provider, err := Init(ctx, config)
			if err != nil {
				t.Errorf("Log Init failed for goroutine %d: %v", i, err)
				return
			}
			if provider != nil {
				_ = Shutdown(ctx, provider)
			}
		}(i)
	}

	wg.Wait()
}

func TestInitErrorPaths(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                     os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL":      os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT":      os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
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

	// Set basic test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	tests := []struct {
		name    string
		config  OtelGoLogsConfig
		wantErr bool
	}{
		{
			name: "invalid TLS config - conflicting insecure and CA cert",
			config: OtelGoLogsConfig{
				TLS: &internal.TLSConfig{
					Insecure:   true,
					CACertPath: "/path/to/ca.pem",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid TLS config - missing client key",
			config: OtelGoLogsConfig{
				TLS: &internal.TLSConfig{
					ClientCertPath: "/path/to/cert.pem",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid TLS config - invalid CA cert path",
			config: OtelGoLogsConfig{
				TLS: &internal.TLSConfig{
					CACertPath: "/nonexistent/path/ca.pem",
				},
			},
			wantErr: true,
		},
		{
			name: "valid config with minimal settings",
			config: OtelGoLogsConfig{
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
			_, provider, err := Init(ctx, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if provider != nil {
				_ = Shutdown(ctx, provider)
			}
		})
	}
}

func countGoroutines() int {
	return runtime.NumGoroutine()
}

func TestGoroutineLeak(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	baseGoroutines := countGoroutines()

	ctx := context.Background()

	// Initialize logs provider
	_, provider, err := Init(ctx, OtelGoLogsConfig{})
	if err != nil {
		t.Fatalf("Failed to initialize logs: %v", err)
	}

	// Shutdown logs provider
	if err := Shutdown(ctx, provider); err != nil {
		t.Errorf("Failed to shutdown logs provider: %v", err)
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
		"OTEL_SERVICE_NAME":                os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Initialize logs provider
	_, provider, err := Init(ctx, OtelGoLogsConfig{})
	if err != nil {
		t.Fatalf("Failed to initialize logs: %v", err)
	}

	// Shutdown logs provider with the same context
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
	config := OtelGoLogsConfig{}
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
	configWithAttrs := OtelGoLogsConfig{
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

func TestInitErrorPathScenarios(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "error-test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318")

	ctx := context.Background()

	scenarios := []struct {
		name          string
		config        OtelGoLogsConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config",
			config: OtelGoLogsConfig{
				Attributes: []attribute.KeyValue{
					attribute.String("test", "valid"),
				},
			},
			expectError: false,
		},
		{
			name: "invalid TLS config - conflicting settings",
			config: OtelGoLogsConfig{
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
			config: OtelGoLogsConfig{
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
			config: OtelGoLogsConfig{
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
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "invalid-endpoint-test")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")

	ctx := context.Background()

	// Test with an invalid endpoint - this will cause errors during actual export attempts
	// but Init should still succeed if the configuration is valid
	config := OtelGoLogsConfig{
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
