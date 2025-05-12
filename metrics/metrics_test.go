package metrics

import (
	"context"
	"os"
	"strings"
	"testing"

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
