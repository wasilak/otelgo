package tracing

import (
	"context"
	"log"
	"os"
	"time"

	"dario.cat/mergo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// The OtelGoTracingConfig type is used to configure whether host metrics are enabled or not.
// @property {bool} HostMetricsEnabled - A boolean value that indicates whether host metrics are
// enabled or not.
type OtelGoTracingConfig struct {
	HostMetricsEnabled bool
}

var defaultConfig = OtelGoTracingConfig{
	HostMetricsEnabled: false,
}

// The `InitTracer` function initializes an OpenTelemetry tracer with a specified configuration,
// exporter, and resource.
func InitTracer(ctx context.Context, config OtelGoTracingConfig) {

	// The `mergo.Merge(&defaultConfig, config)` statement is merging the values of the `config` variable
	// into the `defaultConfig` variable.
	mergo.Merge(&defaultConfig, config)

	var err error
	var client otlptrace.Client

	// This code block is checking the value of the environment variable `OTEL_EXPORTER_OTLP_PROTOCOL`. If
	// the value is "grpc", it creates a new gRPC client using `otlptracegrpc.NewClient()`. If the value
	// is anything else, it creates a new HTTP client using `otlptracehttp.NewClient()`. The client is
	// used later to initialize the exporter for OpenTelemetry tracing.
	if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc" {
		client = otlptracegrpc.NewClient()
	} else {
		client = otlptracehttp.NewClient()
	}

	// The code `exporter, err := otlptrace.New(ctx, client)` is initializing an exporter for
	// OpenTelemetry tracing. It creates a new exporter using the provided client, which can be either a
	// gRPC client (`otlptracegrpc.Client`) or an HTTP client (`otlptracehttp.Client`) based on the value
	// of the environment variable `OTEL_EXPORTER_OTLP_PROTOCOL`.
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %e", err)
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
	)
	if err != nil {
		log.Fatalf("failed to initialize resource: %e", err)
	}

	// The `if defaultConfig.HostMetricsEnabled` condition checks if the `HostMetricsEnabled` field in the
	// `defaultConfig` variable is set to `true`. If it is `true`, it means that host metrics are enabled.
	if defaultConfig.HostMetricsEnabled {
		interval := 2 * time.Second
		setupHostMetrics(ctx, res, interval)
	}

	// Create the trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set the global trace provider
	otel.SetTracerProvider(tp)

	// Set the propagator
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagator)
}
