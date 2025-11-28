package logs

import (
	"context"
	"os"
	"testing"

	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/otel/attribute"
)

func TestLogsBuilder(t *testing.T) {
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
		name        string
		builderFunc func() *LogsBuilder
	}{
		{
			name: "basic builder",
			builderFunc: func() *LogsBuilder {
				return NewBuilder()
			},
		},
		{
			name: "builder with attributes",
			builderFunc: func() *LogsBuilder {
				return NewBuilder().
					WithAttributes(attribute.String("env", "test")).
					WithAttributes(attribute.Int("version", 1))
			},
		},
		{
			name: "builder with TLS",
			builderFunc: func() *LogsBuilder {
				return NewBuilder().
					WithTLS(&internal.TLSConfig{Insecure: true}).
					WithAttributes(attribute.String("secure", "false"))
			},
		},
		{
			name: "builder with TLS insecure",
			builderFunc: func() *LogsBuilder {
				return NewBuilder().
					WithTLSInsecure()
			},
		},
		{
			name: "builder with custom TLS",
			builderFunc: func() *LogsBuilder {
				return NewBuilder().
					WithTLSCustom(true, "", "", "", "example.com")
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

func TestLogsBuilderFluentAPI(t *testing.T) {
	// Test that the fluent API works as expected
	builder := NewBuilder()

	// Verify we can chain calls
	builder = builder.
		WithAttributes(attribute.String("key1", "value1")).
		WithAttributes(attribute.String("key2", "value2")).
		WithTLSInsecure()

	// Check internal state is properly configured
	if builder.config.TLS == nil || !builder.config.TLS.Insecure {
		t.Error("TLS insecure should be set")
	}

	if len(builder.config.Attributes) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(builder.config.Attributes))
	}
}

func TestLogsBuilderWithServiceName(t *testing.T) {
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
