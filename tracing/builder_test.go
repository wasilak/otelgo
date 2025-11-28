package tracing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/wasilak/otelgo/internal"
)

func TestTracingBuilder(t *testing.T) {
	// Save original env vars and restore them after test
	originalVars := map[string]string{
		"OTEL_SERVICE_NAME":                  os.Getenv("OTEL_SERVICE_NAME"),
		"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL": os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"),
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"),
	}
	defer func() {
		for k, v := range originalVars {
			os.Setenv(k, v)
		}
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://localhost:4318")

	tests := []struct {
		name        string
		builderFunc func() *TracingBuilder
	}{
		{
			name: "basic builder",
			builderFunc: func() *TracingBuilder {
				return NewBuilder()
			},
		},
		{
			name: "builder with host metrics enabled",
			builderFunc: func() *TracingBuilder {
				return NewBuilder().
					WithHostMetrics(true, 5*time.Second)
			},
		},
		{
			name: "builder with runtime metrics enabled",
			builderFunc: func() *TracingBuilder {
				return NewBuilder().
					WithRuntimeMetrics(true, 10*time.Second)
			},
		},
		{
			name: "builder with host and runtime metrics enabled",
			builderFunc: func() *TracingBuilder {
				return NewBuilder().
					WithHostMetrics(true, 3*time.Second).
					WithRuntimeMetrics(true, 7*time.Second)
			},
		},
		{
			name: "builder with TLS",
			builderFunc: func() *TracingBuilder {
				return NewBuilder().
					WithTLS(&internal.TLSConfig{Insecure: true}).
					WithHostMetrics(true, 5*time.Second)
			},
		},
		{
			name: "builder with TLS insecure",
			builderFunc: func() *TracingBuilder {
				return NewBuilder().
					WithTLSInsecure()
			},
		},
		{
			name: "builder with custom TLS",
			builderFunc: func() *TracingBuilder {
				return NewBuilder().
					WithTLSCustom(true, "", "", "", "example.com")
			},
		},
		{
			name: "builder with all configurations",
			builderFunc: func() *TracingBuilder {
				return NewBuilder().
					WithHostMetrics(true, 3*time.Second).
					WithRuntimeMetrics(true, 5*time.Second).
					WithTLSInsecure()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			builder := tt.builderFunc()

			newCtx, provider, err := builder.Build(ctx)
			if err != nil {
				// For this test, we expect potential errors due to connection attempts
				// The important thing is that the builder functions work without panicking
				t.Logf("Build returned error (expected): %v", err)
			}

			if newCtx == nil {
				t.Error("Context should not be nil")
			}

			if provider != nil {
				// Only try to shutdown if provider was created successfully
				_ = Shutdown(newCtx, provider)
			}
		})
	}
}

func TestTracingBuilderFluentAPI(t *testing.T) {
	// Test that the fluent API works as expected
	builder := NewBuilder()

	// Verify we can chain calls
	builder = builder.
		WithHostMetrics(true, 10*time.Second).
		WithRuntimeMetrics(true, 15*time.Second).
		WithTLSInsecure()

	// Check internal state is properly configured
	if !builder.config.HostMetricsEnabled || builder.config.HostMetricsInterval != 10*time.Second {
		t.Error("Host metrics configuration should be set correctly")
	}

	if !builder.config.RuntimeMetricsEnabled || builder.config.RuntimeMetricsInterval != 15*time.Second {
		t.Error("Runtime metrics configuration should be set correctly")
	}

	if builder.config.TLS == nil || !builder.config.TLS.Insecure {
		t.Error("TLS insecure should be set")
	}
}

func TestTracingBuilderWithServiceName(t *testing.T) {
	// Save original env vars and restore them after test
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	defer os.Setenv("OTEL_SERVICE_NAME", originalServiceName)

	os.Setenv("OTEL_SERVICE_NAME", "test-builder-service")

	ctx := context.Background()

	// Test with basic config
	_, provider, err := NewBuilder().Build(ctx)
	if err != nil {
		t.Logf("Build returned error (expected): %v", err)
	}

	if provider != nil {
		_ = Shutdown(ctx, provider)
	}
}
