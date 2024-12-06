package logs

import (
	"context"

	"dario.cat/mergo"
	"github.com/wasilak/otelgo/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// OtelGoLogsConfig specifies the configuration for the OpenTelemetry logs.
type OtelGoLogsConfig struct {
	Attributes []attribute.KeyValue `json:"attributes"` // Attributes specifies the attributes to be added to the logger resource. Default is an empty slice.
}

// defaultConfig specifies the default configuration for the OpenTelemetry logs.
var defaultConfig = OtelGoLogsConfig{
	Attributes: []attribute.KeyValue{},
}

// Init initializes an OpenTelemetry logger with a specified configuration.
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

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL") {
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

	global.SetLoggerProvider(logProvider)

	return ctx, logProvider, nil
}

// Shutdown closes the logger provider.
func Shutdown(ctx context.Context, logProvider *sdk.LoggerProvider) {
	defer func() {
		if err := logProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
}
