package tracing

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	// Save original env vars and restore them after test
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	// Set test environment
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "basic initialization",
			config: Config{
				HostMetricsEnabled:     false,
				RuntimeMetricsEnabled:  false,
				HostMetricsInterval:    2 * time.Second,
				RuntimeMetricsInterval: 2 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "with host metrics enabled",
			config: Config{
				HostMetricsEnabled:     true,
				RuntimeMetricsEnabled:  false,
				HostMetricsInterval:    2 * time.Second,
				RuntimeMetricsInterval: 2 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "with runtime metrics enabled",
			config: Config{
				HostMetricsEnabled:     false,
				RuntimeMetricsEnabled:  true,
				HostMetricsInterval:    2 * time.Second,
				RuntimeMetricsInterval: 2 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "with both metrics enabled",
			config: Config{
				HostMetricsEnabled:     true,
				RuntimeMetricsEnabled:  true,
				HostMetricsInterval:    2 * time.Second,
				RuntimeMetricsInterval: 2 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "with custom intervals",
			config: Config{
				HostMetricsEnabled:     true,
				RuntimeMetricsEnabled:  true,
				HostMetricsInterval:    5 * time.Second,
				RuntimeMetricsInterval: 10 * time.Second,
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
	config := Config{
		HostMetricsEnabled:     false,
		RuntimeMetricsEnabled:  false,
		HostMetricsInterval:    2 * time.Second,
		RuntimeMetricsInterval: 2 * time.Second,
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
