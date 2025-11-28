package metrics

import (
	"context"

	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/otel/attribute"
	sdk "go.opentelemetry.io/otel/sdk/metric"
)

// MetricsBuilder provides a fluent API for configuring OpenTelemetry metrics.
type MetricsBuilder struct {
	config OtelGoMetricsConfig
}

// NewBuilder creates a new MetricsBuilder with default configuration.
func NewBuilder() *MetricsBuilder {
	return &MetricsBuilder{
		config: OtelGoMetricsConfig{
			Attributes: []attribute.KeyValue{},
			TLS:        nil,
		},
	}
}

// WithAttributes sets the attributes to be added to the metric resource.
// This method can be called multiple times to add more attributes.
func (b *MetricsBuilder) WithAttributes(attrs ...attribute.KeyValue) *MetricsBuilder {
	b.config.Attributes = append(b.config.Attributes, attrs...)
	return b
}

// WithTLS sets the TLS configuration for exporters.
func (b *MetricsBuilder) WithTLS(tlsConfig *internal.TLSConfig) *MetricsBuilder {
	b.config.TLS = tlsConfig
	return b
}

// WithTLSInsecure configures the builder to skip TLS verification (insecure mode).
func (b *MetricsBuilder) WithTLSInsecure() *MetricsBuilder {
	b.config.TLS = &internal.TLSConfig{
		Insecure: true,
	}
	return b
}

// WithTLSCustom configures the builder with custom TLS settings.
func (b *MetricsBuilder) WithTLSCustom(insecure bool, caCertPath, clientCertPath, clientKeyPath, serverName string) *MetricsBuilder {
	b.config.TLS = &internal.TLSConfig{
		Insecure:       insecure,
		CACertPath:     caCertPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		ServerName:     serverName,
	}
	return b
}

// Build constructs the OtelGoMetricsConfig and initializes the OpenTelemetry metric provider.
// It validates the configuration and returns the context, provider, and any error encountered.
func (b *MetricsBuilder) Build(ctx context.Context) (context.Context, *sdk.MeterProvider, error) {
	return Init(ctx, b.config)
}
