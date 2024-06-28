package tracing

import (
	"context"
	"log"
	"time"

	"github.com/wasilak/otelgo/common"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func setupRuntimeMetrics(ctx context.Context, res *resource.Resource, interval time.Duration) {
	var err error
	var exp metric.Exporter

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL") {
		exp, err = otlpmetricgrpc.New(ctx)
	} else {
		exp, err = otlpmetrichttp.New(ctx)
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
