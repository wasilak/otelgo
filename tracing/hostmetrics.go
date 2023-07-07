package tracing

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// The function sets up host metrics collection using the OpenTelemetry exporter based on the specified
// protocol (gRPC or HTTP).
func setupHostMetrics(ctx context.Context, res *resource.Resource, interval time.Duration) {
	var err error
	var exp metric.Exporter

	// This code block is checking the value of the environment variable `OTEL_EXPORTER_OTLP_PROTOCOL`. If
	// the value is "grpc", it creates a new gRPC exporter using `otlpmetricgrpc.New(ctx)`. If the value
	// is not "grpc", it creates a new HTTP exporter using `otlpmetrichttp.New(ctx)`.
	if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc" {
		exp, err = otlpmetricgrpc.New(ctx)
	} else {
		exp, err = otlpmetrichttp.New(ctx)
	}
	if err != nil {
		panic(err)
	}

	// The code block is setting up the periodic collection of host metrics using OpenTelemetry.
	read := metric.NewPeriodicReader(exp, metric.WithInterval(interval))
	provider := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(read))
	defer func() {
		err := provider.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Starting host instrumentation")

	// The code `err = host.Start(host.WithMeterProvider(provider))` is starting the host instrumentation
	// for collecting host metrics using OpenTelemetry. It uses the `host.Start` function from the
	// OpenTelemetry `host` package and passes the `provider` as the meter provider for collecting
	// metrics.
	err = host.Start(host.WithMeterProvider(provider))
	if err != nil {
		log.Fatal(err)
	}
}
