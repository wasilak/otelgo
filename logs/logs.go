// Shutdown gracefully shuts down the logger provider, ensuring all logs are flushed.
//
// ✅ Proper usage pattern:
//   1. Initialize the logger via Init()
//   2. Defer Shutdown() with the provider as an argument
//   3. Ensure the context is valid for the duration of the program
//
// Example:
//   ctx := context.Background()
//   provider, _ := logs.Init(ctx, logs.OtelGoLogsConfig{
//       Attributes: []attribute.KeyValue{
//           semconv.ServiceNameKey.String("my-service"),
//           semconv.ServiceVersionKey.String("1.0.0"),
//       },
//   })
//   defer logs.Shutdown(ctx, provider) // Ensures shutdown happens on exit
//
// ⚠️ Critical: Always call Shutdown() when your application is exiting to prevent data loss
//              Defer ensures this happens even if the program terminates unexpectedly

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

// Init initializes the OpenTelemetry logger provider with the specified configuration.
// It sets up a log pipeline by configuring exporters and resource attributes.
//
// The function automatically merges provided configuration with defaults and sets up
// appropriate OTLP exporters based on the environment configuration.
//
// Parameters:
//   - ctx: The context for controlling logger initialization lifetime
//   - config: The configuration containing logger setup options and attributes
//
// Returns:
//   - context.Context: Updated context with logger provider
//   - *sdk.LoggerProvider: Configured logger provider for emitting logs
//   - error: Non-nil if initialization fails
//
// Example:
//
//	config := logs.OtelGoLogsConfig{
//	    Attributes: []attribute.KeyValue{
//	        semconv.ServiceNameKey.String("my-service"),
//	    },
//	}
//	ctx, provider, err := logs.Init(context.Background(), config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer func() {
//	    if err := provider.Shutdown(ctx); err != nil {
//	        log.Printf("failed to shutdown provider: %v", err)
//	    }
//	}()
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

// Shutdown gracefully shuts down the logger provider and flushes any pending logs.
// It should be called when the application is terminating to ensure all logs are exported.
//
// Parameters:
//   - ctx: The context for controlling shutdown timeout
//   - logProvider: The provider instance to shut down
//
// Returns:
//   - error: Non-nil if shutdown fails
//
// Example:
//
//	ctx := context.Background()
//	ctx, provider, err := logs.Init(ctx, logs.OtelGoLogsConfig{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer logs.Shutdown(ctx, provider)
func Shutdown(ctx context.Context, logProvider *sdk.LoggerProvider) error {
	return logProvider.Shutdown(ctx)
}
