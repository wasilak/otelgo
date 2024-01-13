package metrics

import (
	"context"
	"os"

	"dario.cat/mergo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// var requestCount metric.Int64Counter

type OtelGoMetricsConfig struct {
	Attributes []attribute.KeyValue
}

var defaultConfig = OtelGoMetricsConfig{
	Attributes: []attribute.KeyValue{
		semconv.ServiceNameKey.String(os.Getenv("OTEL_SERVICE_NAME")),
		semconv.ServiceVersionKey.String("v0.0.0"),
	},
}

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

	// Instantiate the OTLP HTTP exporter
	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return ctx, nil, err
	}

	// Instantiate the OTLP HTTP exporter
	meterProvider := sdk.NewMeterProvider(
		sdk.WithResource(res),
		sdk.WithReader(sdk.NewPeriodicReader(exporter)),
	)

	// defer func() {
	// 	err := meterProvider.Shutdown(context.Background())
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// }()

	otel.SetMeterProvider(meterProvider)

	return ctx, meterProvider, nil
}
