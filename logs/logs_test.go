package logs

import (
	"context"
	"os"
	"testing"

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
