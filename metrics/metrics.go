package metrics

import (
	"context"
	"os"

	"dario.cat/mergo"
	"github.com/wasilak/otelgo/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// OtelGoMetricsConfig specifies the configuration for the OpenTelemetry metrics.
type OtelGoMetricsConfig struct {
	Attributes []attribute.KeyValue `json:"attributes"` // Attributes specifies the attributes to be added to the metric resource. Default is an empty slice.
}

// defaultConfig specifies the default configuration for the OpenTelemetry metrics.
var defaultConfig = OtelGoMetricsConfig{
	Attributes: []attribute.KeyValue{
		semconv.ServiceNameKey.String(os.Getenv("OTEL_SERVICE_NAME")),
		semconv.ServiceVersionKey.String("v0.0.0"),
	},
}

// Init initializes the OpenTelemetry metric provider with the specified configuration.
// It sets up a federated metric pipeline by configuring exporters and resource attributes.
//
// The function automatically merges provided configuration with defaults and sets up
// appropriate OTLP exporters based on the environment configuration.
//
// Parameters:
//   - ctx: The context for controlling metric initialization lifetime
//   - config: The configuration containing metric setup options and attributes
//
// Returns:
//   - context.Context: Updated context with metric provider
//   - *sdk.MeterProvider: Configured metric provider for emitting metrics
//   - error: Non-nil if initialization fails
//
// Example:
//
//	config := metrics.OtelGoMetricsConfig{
//	    Attributes: []attribute.KeyValue{
//	        semconv.ServiceNameKey.String("my-service"),
//	    },
//	}
//	ctx, provider, err := metrics.Init(context.Background(), config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer func() {
//	    if err := provider.Shutdown(ctx); err != nil {
//	        log.Printf("failed to shutdown provider: %v", err)
//	    }
//	}()
func Init(ctx context.Context, config OtelGoMetricsConfig) (context.Context, *sdk.MeterProvider, error) {
	err := mergo.Merge(&defaultConfig, config, mergo.WithOverride)
	if err != nil {
		return ctx, nil, err
	}

	res, err := resource.New(ctx,
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithFromEnv(),
		resource.WithAttributes(defaultConfig.Attributes...),
	)
	if err != nil {
		return ctx, nil, err
	}

	var exporter sdk.Exporter

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL") {
		exporter, err = otlpmetricgrpc.New(ctx)
		if err != nil {
			return ctx, nil, err
		}
	} else {
		exporter, err = otlpmetrichttp.New(ctx)
		if err != nil {
			return ctx, nil, err
		}
	}

	meterProvider := sdk.NewMeterProvider(
		sdk.WithResource(res),
		sdk.WithReader(sdk.NewPeriodicReader(exporter)),
	)

	otel.SetMeterProvider(meterProvider)

	return ctx, meterProvider, nil
}

// Shutdown gracefully shuts down the metric provider and flushes any pending metrics.
// It should be called when the application is terminating to ensure all metrics are exported.
//
// Parameters:
//   - ctx: The context for controlling shutdown timeout
//   - meterProvider: The provider instance to shut down
//
// Returns:
//   - error: Non-nil if shutdown fails
//
// Example:
//
//	ctx := context.Background()
//	ctx, provider, err := metrics.Init(ctx, metrics.OtelGoMetricsConfig{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer metrics.Shutdown(ctx, provider)
func Shutdown(ctx context.Context, meterProvider *sdk.MeterProvider) error {
	return meterProvider.Shutdown(ctx)
}
