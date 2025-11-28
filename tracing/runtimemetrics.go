package tracing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wasilak/otelgo/common"
	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// setupRuntimeMetrics configures and starts the runtime metrics collection with the specified settings.
// It initializes a metric exporter based on the configured protocol (gRPC or HTTP) and sets up
// periodic collection of Go runtime metrics (e.g., memory stats, goroutine count).
//
// Parameters:
//   - ctx: The context for controlling the metrics setup lifetime
//   - res: The resource to associate with the metrics
//   - interval: The duration between metric collections
//
// The function will panic if it fails to create the exporter or start the runtime metrics collection.
// This is intentional as runtime metrics are critical for monitoring and the application should not
// continue without them if they were explicitly enabled.
func setupRuntimeMetrics(ctx context.Context, res *resource.Resource, interval time.Duration) {
	var err error
	var exp metric.Exporter

	// Use the same TLS configuration as other components
	tlsConfigInternal := internal.NewTLSConfig()
	if err := tlsConfigInternal.Validate(); err != nil {
		panic(fmt.Errorf("invalid TLS configuration for runtime metrics: %w", err))
	}

	tlsConfig, err := tlsConfigInternal.BuildTLSConfig()
	if err != nil {
		panic(fmt.Errorf("failed to build TLS config for runtime metrics: %w", err))
	}

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL") {
		// Configure gRPC dial options to use the custom TLS configuration
		grpcOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		}

		exp, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithDialOption(grpcOpts...))
	} else {
		exp, err = otlpmetrichttp.New(ctx, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
	}
	if err != nil {
		panic(err)
	}

	read := metric.NewPeriodicReader(exp, metric.WithInterval(interval))
	provider := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(read))

	err = runtime.Start(runtime.WithMeterProvider(provider))
	if err != nil {
		log.Fatal(err)
	}
}
