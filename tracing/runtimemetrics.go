package tracing

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/wasilak/otelgo/common"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func setupRuntimeMetrics(ctx context.Context, res *resource.Resource, interval time.Duration) {
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

	err = runtime.Start(runtime.WithMeterProvider(provider))
	if err != nil {
		log.Fatal(err)
	}
}
