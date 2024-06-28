package tracing

import (
	"context"
	"time"

	"dario.cat/mergo"
	"github.com/wasilak/otelgo/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
)

// The OtelGoTracingConfig type is used to configure whether host metrics are enabled or not.
// @property {bool} HostMetricsEnabled - A boolean value that indicates whether host metrics are
// enabled or not.
type OtelGoTracingConfig struct {
	HostMetricsEnabled     bool          `json:"host_metrics_enabled"`     // HostMetricsEnabled specifies whether host metrics are enabled. Default is false.
	HostMetricsInterval    time.Duration `json:"host_metrics_interval"`    // HostMetricsInterval specifies the interval at which host metrics are collected. Default is 2 seconds.
	RuntimeMetricsEnabled  bool          `json:"runtime_metrics_enabled"`  // RuntimeMetricsEnabled specifies whether runtime metrics are enabled. Default is false.
	RuntimeMetricsInterval time.Duration `json:"runtime_metrics_interval"` // RuntimeMetricsInterval specifies the interval at which runtime metrics are collected. Default is 2 seconds.
}

var defaultConfig = OtelGoTracingConfig{
	HostMetricsEnabled:     false,
	RuntimeMetricsEnabled:  false,
	HostMetricsInterval:    2 * time.Second,
	RuntimeMetricsInterval: 2 * time.Second,
}

// The `InitTracer` function initializes an OpenTelemetry tracer with a specified configuration,
// exporter, and resource.
func Init(ctx context.Context, config OtelGoTracingConfig) (context.Context, *sdk.TracerProvider, error) {

	// The code `err := mergo.Merge(&defaultConfig, config, mergo.WithOverride)` is using the `mergo`
	// library to merge the `config` object into the `defaultConfig` object.
	err := mergo.Merge(&defaultConfig, config, mergo.WithOverride)
	if err != nil {
		return ctx, nil, err
	}

	var client otlptrace.Client

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL") {
		client = otlptracegrpc.NewClient()
	} else {
		client = otlptracehttp.NewClient()
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
	)
	if err != nil {
		return ctx, nil, err
	}

	// The `if defaultConfig.HostMetricsEnabled` condition checks if the `HostMetricsEnabled` field in the
	// `defaultConfig` variable is set to `true`. If it is `true`, it means that host metrics are enabled.
	if defaultConfig.HostMetricsEnabled {
		setupHostMetrics(ctx, res, defaultConfig.HostMetricsInterval.Abs())
	}

	if defaultConfig.RuntimeMetricsEnabled {
		setupRuntimeMetrics(ctx, res, defaultConfig.RuntimeMetricsInterval)
	}

	// Create the trace provider
	tp := sdk.NewTracerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithResource(res),
	)

	// Set the global trace provider
	otel.SetTracerProvider(tp)

	// Set the propagator
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagator)

	return ctx, tp, nil
}
