package tracing

import (
	"context"
	"time"

	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/otel/sdk/trace"
)

// TracingBuilder provides a fluent API for configuring OpenTelemetry tracing.
type TracingBuilder struct {
	config Config
}

// NewBuilder creates a new TracingBuilder with default configuration.
func NewBuilder() *TracingBuilder {
	return &TracingBuilder{
		config: Config{
			HostMetricsEnabled:     false,
			HostMetricsInterval:    2 * time.Second,
			RuntimeMetricsEnabled:  false,
			RuntimeMetricsInterval: 2 * time.Second,
			TLS:                    nil,
		},
	}
}

// WithHostMetrics enables host metrics collection with the specified interval.
func (b *TracingBuilder) WithHostMetrics(enabled bool, interval time.Duration) *TracingBuilder {
	b.config.HostMetricsEnabled = enabled
	if interval > 0 {
		b.config.HostMetricsInterval = interval
	}
	return b
}

// WithRuntimeMetrics enables runtime metrics collection with the specified interval.
func (b *TracingBuilder) WithRuntimeMetrics(enabled bool, interval time.Duration) *TracingBuilder {
	b.config.RuntimeMetricsEnabled = enabled
	if interval > 0 {
		b.config.RuntimeMetricsInterval = interval
	}
	return b
}

// WithTLS sets the TLS configuration for exporters.
func (b *TracingBuilder) WithTLS(tlsConfig *internal.TLSConfig) *TracingBuilder {
	b.config.TLS = tlsConfig
	return b
}

// WithTLSInsecure configures the builder to skip TLS verification (insecure mode).
func (b *TracingBuilder) WithTLSInsecure() *TracingBuilder {
	b.config.TLS = &internal.TLSConfig{
		Insecure: true,
	}
	return b
}

// WithTLSCustom configures the builder with custom TLS settings.
func (b *TracingBuilder) WithTLSCustom(insecure bool, caCertPath, clientCertPath, clientKeyPath, serverName string) *TracingBuilder {
	b.config.TLS = &internal.TLSConfig{
		Insecure:       insecure,
		CACertPath:     caCertPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		ServerName:     serverName,
	}
	return b
}

// Build constructs the Config and initializes the OpenTelemetry tracer provider.
// It validates the configuration and returns the context, provider, and any error encountered.
func (b *TracingBuilder) Build(ctx context.Context) (context.Context, *trace.TracerProvider, error) {
	return Init(ctx, b.config)
}
