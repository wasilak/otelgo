package tracing

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/wasilak/otelgo/common"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// setupHostMetrics configures and starts the host metrics collection with the specified settings.
// It initializes a metric exporter based on the configured protocol (gRPC or HTTP) and sets up
// periodic collection of host-level metrics.
//
// Parameters:
//   - ctx: The context for controlling the metrics setup lifetime
//   - res: The resource to associate with the metrics
//   - interval: The duration between metric collections
//
// The function will panic if it fails to create the exporter or start the host metrics collection.
// This is intentional as host metrics are critical for monitoring and the application should not
// continue without them if they were explicitly enabled.
func setupHostMetrics(ctx context.Context, res *resource.Resource, interval time.Duration) {
	var err error
	var exp metric.Exporter

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL") {
		// Create a custom TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // WARNING: This skips certificate verification!
		}

		// Configure gRPC dial options to use the custom TLS configuration
		grpcOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		}

		exp, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithDialOption(grpcOpts...))
	} else {
		// Create a custom TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // WARNING: Skips certificate verification
		}
		exp, err = otlpmetrichttp.New(ctx, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
	}
	if err != nil {
		panic(err)
	}

	read := metric.NewPeriodicReader(exp, metric.WithInterval(interval))
	provider := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(read))

	err = host.Start(host.WithMeterProvider(provider))
	if err != nil {
		log.Fatal(err)
	}
}
