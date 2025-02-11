package logs

import (
	"context"
	"crypto/tls"
	"os"

	"dario.cat/mergo"
	"github.com/wasilak/otelgo/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// OtelGoLogsConfig specifies the configuration for the OpenTelemetry logs.
type OtelGoLogsConfig struct {
	Attributes []attribute.KeyValue `json:"attributes"` // Attributes specifies the attributes to be added to the logger resource. Default is an empty slice.
}

// defaultConfig specifies the default configuration for the OpenTelemetry logs.
var defaultConfig = OtelGoLogsConfig{
	Attributes: []attribute.KeyValue{
		semconv.ServiceNameKey.String(os.Getenv("OTEL_SERVICE_NAME")),
		semconv.ServiceVersionKey.String("v0.0.0"),
	},
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

	if common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL") { // Create a custom TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // WARNING: This skips certificate verification!
		}

		// Configure gRPC dial options to use the custom TLS configuration
		grpcOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		}

		exporter, err = otlploggrpc.New(ctx, otlploggrpc.WithDialOption(grpcOpts...))
		if err != nil {
			return ctx, nil, err
		}
	} else {
		// Create a custom TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // WARNING: Skips certificate verification
		}
		exporter, err = otlploghttp.New(ctx, otlploghttp.WithTLSClientConfig(tlsConfig))
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
