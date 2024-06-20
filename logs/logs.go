package logs

import (
	"context"
	"log"
	"os"

	"dario.cat/mergo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// var requestCount metric.Int64Counter

type OtelGoLogsConfig struct {
	Attributes []attribute.KeyValue
}

var defaultConfig = OtelGoLogsConfig{
	Attributes: []attribute.KeyValue{
		semconv.ServiceNameKey.String(os.Getenv("OTEL_SERVICE_NAME")),
		semconv.ServiceVersionKey.String("v0.0.0"),
	},
}

func Init(ctx context.Context, config OtelGoLogsConfig) (context.Context, *sdk.LoggerProvider, error) {
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

	var exporter sdk.Exporter

	if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc" {
		exporter, err = otlploggrpc.New(ctx)
		if err != nil {
			return ctx, nil, err
		}
	} else {
		exporter, err = otlploghttp.New(ctx)
		if err != nil {
			return ctx, nil, err
		}
	}

	processor := sdk.NewBatchProcessor(exporter)

	logProvider := sdk.NewLoggerProvider(
		sdk.WithResource(res),
		sdk.WithProcessor(processor),
	)

	// Handle shutdown properly so nothing leaks.
	defer func() {
		if err := logProvider.Shutdown(ctx); err != nil {
			log.Fatalln(err)
		}
	}()

	global.SetLoggerProvider(logProvider)

	return ctx, logProvider, nil
}
