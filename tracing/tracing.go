package tracing

import (
	"context"
	"fmt"
	"time"

	"dario.cat/mergo"
	"github.com/wasilak/otelgo/common"
	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Sampler control
// https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#general-sdk-configuration
// OTEL_TRACES_SAMPLER see: https://opentelemetry.io/docs/specs/otel/trace/sdk/#sampling

// The Config type is used to configure whether host metrics are enabled or not.
// @property {bool} HostMetricsEnabled - A boolean value that indicates whether host metrics are
// enabled or not.
type Config struct {
	HostMetricsEnabled     bool                `json:"host_metrics_enabled"`     // HostMetricsEnabled specifies whether host metrics are enabled. Default is false.
	HostMetricsInterval    time.Duration       `json:"host_metrics_interval"`    // HostMetricsInterval specifies the interval at which host metrics are collected. Default is 2 seconds.
	RuntimeMetricsEnabled  bool                `json:"runtime_metrics_enabled"`  // RuntimeMetricsEnabled specifies whether runtime metrics are enabled. Default is false.
	RuntimeMetricsInterval time.Duration       `json:"runtime_metrics_interval"` // RuntimeMetricsInterval specifies the interval at which runtime metrics are collected. Default is 2 seconds.
	TLS                    *internal.TLSConfig // TLS specifies the TLS configuration for exporters. Default is nil.
}

// The defaultConfig variable is an instance of the Config struct that specifies the default configuration
var defaultConfig = Config{
	HostMetricsEnabled:     false,
	RuntimeMetricsEnabled:  false,
	HostMetricsInterval:    2 * time.Second,
	RuntimeMetricsInterval: 2 * time.Second,
}

// Init initializes the OpenTelemetry tracer provider with the specified configuration.
// It sets up a trace pipeline by configuring exporters and resource attributes.
//
// The function automatically merges provided configuration with defaults and sets up
// appropriate OTLP exporters based on the environment configuration. It also configures
// host and runtime metrics if enabled in the configuration.
//
// Parameters:
//   - ctx: The context for controlling tracer initialization lifetime
//   - config: The configuration containing tracer setup options and metrics settings
//
// Returns:
//   - context.Context: Updated context with tracer provider
//   - *trace.TracerProvider: Configured tracer provider for creating spans
//   - error: Non-nil if initialization fails
//
// Example:
//
//	config := tracing.Config{
//	    HostMetricsEnabled: true,
//	    RuntimeMetricsEnabled: true,
//	    HostMetricsInterval: 5 * time.Second,
//	}
//	ctx, provider, err := tracing.Init(context.Background(), config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer func() {
//	    if err := provider.Shutdown(ctx); err != nil {
//	        log.Printf("failed to shutdown provider: %v", err)
//	    }
//	}()
func Init(ctx context.Context, config Config) (context.Context, *trace.TracerProvider, error) {
	localConfig := Config{
		HostMetricsEnabled:     defaultConfig.HostMetricsEnabled,
		HostMetricsInterval:    defaultConfig.HostMetricsInterval,
		RuntimeMetricsEnabled:  defaultConfig.RuntimeMetricsEnabled,
		RuntimeMetricsInterval: defaultConfig.RuntimeMetricsInterval,
		TLS:                    config.TLS,
	}

	// The code `err := mergo.Merge(&defaultConfig, config, mergo.WithOverride)` is using the `mergo`
	// library to merge the `config` object into the `defaultConfig` object.
	err := mergo.Merge(&localConfig, config, mergo.WithOverride)
	if err != nil {
		return ctx, nil, err
	}

	if localConfig.TLS == nil {
		localConfig.TLS = internal.NewTLSConfig()
	}

	if err := localConfig.TLS.Validate(); err != nil {
		return ctx, nil, fmt.Errorf("invalid TLS configuration: %w", err)
	}

	tlsConfig, err := localConfig.TLS.BuildTLSConfig()
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to build TLS config: %w", err)
	}

	var client otlptrace.Client

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL") {
		// Configure gRPC dial options to use the custom TLS configuration
		grpcOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		}

		client = otlptracegrpc.NewClient(
			otlptracegrpc.WithDialOption(grpcOpts...),
		)
	} else {
		client = otlptracehttp.NewClient(
			otlptracehttp.WithTLSClientConfig(tlsConfig),
		)
	}

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return ctx, nil, err
	}

	// The code block is initializing a resource for OpenTelemetry tracing. The `resource.New()` function
	// is called with a context and a series of options (`resource.WithHost()`,
	// `resource.WithContainer()`, etc.) to configure the resource. These options specify the attributes
	// of the resource, such as the host, container, process, telemetry SDK, operating system, and
	// environment variables.
	res, err := resource.New(ctx,
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithFromEnv(),
		resource.WithAttributes(attribute.String("service.version", "v1.0.0")), // Default to v1.0.0 for this release
	)
	if err != nil {
		return ctx, nil, err
	}

	// The `if defaultConfig.HostMetricsEnabled` condition checks if the `HostMetricsEnabled` field in the
	// `defaultConfig` variable is set to `true`. If it is `true`, it means that host metrics are enabled.
	if defaultConfig.HostMetricsEnabled {
		setupHostMetrics(ctx, res, defaultConfig.HostMetricsInterval)
	}

	if defaultConfig.RuntimeMetricsEnabled {
		setupRuntimeMetrics(ctx, res, defaultConfig.RuntimeMetricsInterval)
	}

	// Create the trace provider
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// Set the global trace provider
	otel.SetTracerProvider(traceProvider)

	// Set the propagator
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagator)

	return ctx, traceProvider, nil
}

// Shutdown gracefully shuts down the tracer provider and flushes any pending spans.
// It should be called when the application is terminating to ensure all traces are exported.
//
// Parameters:
//   - ctx: The context for controlling shutdown timeout
//   - traceProvider: The provider instance to shut down
//
// Returns:
//   - error: Non-nil if shutdown fails
//
// Example:
//
//	ctx := context.Background()
//	ctx, provider, err := tracing.Init(ctx, tracing.Config{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tracing.Shutdown(ctx, provider)
func Shutdown(ctx context.Context, traceProvider *trace.TracerProvider) error {
	return traceProvider.Shutdown(ctx)
}
