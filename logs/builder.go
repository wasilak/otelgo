package logs

import (
	"context"

	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/otel/attribute"
	sdk "go.opentelemetry.io/otel/sdk/log"
)

// LogsBuilder provides a fluent API for configuring OpenTelemetry logs.
type LogsBuilder struct {
	config OtelGoLogsConfig
}

// NewBuilder creates a new LogsBuilder with default configuration.
func NewBuilder() *LogsBuilder {
	return &LogsBuilder{
		config: OtelGoLogsConfig{
			Attributes: []attribute.KeyValue{},
			TLS:        nil,
		},
	}
}

// WithAttributes sets the attributes to be added to the logger resource.
// This method can be called multiple times to add more attributes.
func (b *LogsBuilder) WithAttributes(attrs ...attribute.KeyValue) *LogsBuilder {
	b.config.Attributes = append(b.config.Attributes, attrs...)
	return b
}

// WithTLS sets the TLS configuration for exporters.
func (b *LogsBuilder) WithTLS(tlsConfig *internal.TLSConfig) *LogsBuilder {
	b.config.TLS = tlsConfig
	return b
}

// WithTLSInsecure configures the builder to skip TLS verification (insecure mode).
func (b *LogsBuilder) WithTLSInsecure() *LogsBuilder {
	b.config.TLS = &internal.TLSConfig{
		Insecure: true,
	}
	return b
}

// WithTLSCustom configures the builder with custom TLS settings.
func (b *LogsBuilder) WithTLSCustom(insecure bool, caCertPath, clientCertPath, clientKeyPath, serverName string) *LogsBuilder {
	b.config.TLS = &internal.TLSConfig{
		Insecure:       insecure,
		CACertPath:     caCertPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		ServerName:     serverName,
	}
	return b
}

// Build constructs the OtelGoLogsConfig and initializes the OpenTelemetry logger provider.
// It validates the configuration and returns the context, provider, and any error encountered.
func (b *LogsBuilder) Build(ctx context.Context) (context.Context, *sdk.LoggerProvider, error) {
	return Init(ctx, b.config)
}
