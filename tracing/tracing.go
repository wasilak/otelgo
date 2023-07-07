package tracing

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// The `InitTracer` function initializes an exporter and a resource for OpenTelemetry tracing, and sets
// the global trace provider and propagator.
func InitTracer(ctx context.Context, withHostMetrics bool) {

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

	if withHostMetrics {
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
