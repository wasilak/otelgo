# API Reference for otelgo

This document provides comprehensive API documentation for the otelgo library, covering all public types, functions, and methods across all packages.

## Table of Contents

- [Common Types](#common-types)
- [Logs Package](#logs-package)
- [Metrics Package](#metrics-package)
- [Tracing Package](#tracing-package)
- [Slog Package](#slog-package)
- [Internal Package](#internal-package)

## Common Types

### Common Functions

#### `common.IsOtlpProtocolGrpc(dataType string) bool`

Determines if the OTLP protocol is set to gRPC for a given data type. Checks both the specific data type environment variable and the general OTLP protocol setting.

**Parameters:**
- `dataType` (string): The type of data to check (e.g., "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")

**Returns:**
- `bool`: true if the protocol is set to "grpc", false otherwise

**Example:**
```go
isGrpc := common.IsOtlpProtocolGrpc("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
```

## Logs Package

### Public Types

#### `OtelGoLogsConfig`

Configuration for OpenTelemetry logs.

**Fields:**
- `Attributes []attribute.KeyValue`: Attributes specifies the attributes to be added to the logger resource. Default is an empty slice.
- `TLS *internal.TLSConfig`: TLS specifies the TLS configuration for exporters. Default is nil.

#### `OtelGoLogsConfig`

Configuration for OpenTelemetry logs with TLS support.

### Public Functions

#### `Init(ctx context.Context, config OtelGoLogsConfig) (context.Context, *sdk.LoggerProvider, error)`

Initializes the OpenTelemetry logger provider with the specified configuration. It sets up a log pipeline by configuring exporters and resource attributes.

**Parameters:**
- `ctx`: The context for controlling logger initialization lifetime
- `config`: The configuration containing logger setup options and attributes

**Returns:**
- `context.Context`: Updated context with logger provider
- `*sdk.LoggerProvider`: Configured logger provider for emitting logs
- `error`: Non-nil if initialization fails

**Example:**
```go
import (
    "context"
    "go.opentelemetry.io/otel/attribute"
    "github.com/wasilak/otelgo/logs"
)

ctx := context.Background()

config := logs.OtelGoLogsConfig{
    Attributes: []attribute.KeyValue{
        attribute.String("service.name", "my-service"),
    },
}

ctx, provider, err := logs.Init(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer logs.Shutdown(ctx, provider)
```

#### `Shutdown(ctx context.Context, logProvider *sdk.LoggerProvider) error`

Shuts down the logger provider and flushes any pending logs. Should be called when the application is terminating to ensure all logs are exported.

**Parameters:**
- `ctx`: The context for controlling shutdown timeout
- `logProvider`: The provider instance to shut down

**Returns:**
- `error`: Non-nil if shutdown fails

**Example:**
```go
defer func() {
    if err := logs.Shutdown(ctx, logProvider); err != nil {
        log.Printf("failed to shutdown provider: %v", err)
    }
}()
```

### Builder Pattern

#### `NewBuilder() *LogsBuilder`

Creates a new LogsBuilder with default configuration.

**Returns:**
- `*LogsBuilder`: A new builder instance with default configuration

#### `(*LogsBuilder) WithAttributes(attrs ...attribute.KeyValue) *LogsBuilder`

Sets the attributes to be added to the logger resource. This method can be chained with other builder methods.

**Parameters:**
- `attrs ...attribute.KeyValue`: The attributes to add to the logger resource

**Returns:**
- `*LogsBuilder`: The same builder instance for method chaining

#### `(*LogsBuilder) WithTLS(tlsConfig *internal.TLSConfig) *LogsBuilder`

Sets the TLS configuration for the logger exporter. Use this for custom TLS settings.

**Parameters:**
- `tlsConfig *internal.TLSConfig`: The TLS configuration to use

**Returns:**
- `*LogsBuilder`: The same builder instance for method chaining

#### `(*LogsBuilder) WithTLSInsecure() *LogsBuilder`

Configures the builder to skip TLS verification (insecure mode). Use only for development/testing.

**Returns:**
- `*LogsBuilder`: The same builder instance for method chaining

#### `(*LogsBuilder) WithTLSCustom(insecure bool, caCertPath, clientCertPath, clientKeyPath, serverName string) *LogsBuilder`

Configures the builder with custom TLS settings.

**Parameters:**
- `insecure bool`: Whether to skip certificate verification
- `caCertPath string`: Path to the CA certificate file
- `clientCertPath string`: Path to the client certificate file
- `clientKeyPath string`: Path to the client private key file
- `serverName string`: Override the server name for verification

**Returns:**
- `*LogsBuilder`: The same builder instance for method chaining

#### `(*LogsBuilder) Build(ctx context.Context) (context.Context, *sdk.LoggerProvider, error)`

Builds and initializes the OpenTelemetry logger provider with the configured settings.

**Parameters:**
- `ctx context.Context`: The context for controlling initialization

**Returns:**
- `context.Context`: Updated context with logger provider
- `*sdk.LoggerProvider`: Configured logger provider
- `error`: Non-nil if initialization fails

## Metrics Package

### Public Types

#### `OtelGoMetricsConfig`

Configuration for OpenTelemetry metrics.

**Fields:**
- `Attributes []attribute.KeyValue`: Attributes specifies the attributes to be added to the metric resource. Default is an empty slice.
- `TLS *internal.TLSConfig`: TLS specifies the TLS configuration for exporters. Default is nil.

### Public Functions

#### `Init(ctx context.Context, config OtelGoMetricsConfig) (context.Context, *sdk.MeterProvider, error)`

Initializes the OpenTelemetry metric provider with the specified configuration. It sets up a federated metric pipeline by configuring exporters and resource attributes.

**Parameters:**
- `ctx`: The context for controlling metric initialization lifetime
- `config`: The configuration containing metric setup options and attributes

**Returns:**
- `context.Context`: Updated context with metric provider
- `*sdk.MeterProvider`: Configured metric provider for emitting metrics
- `error`: Non-nil if initialization fails

**Example:**
```go
import (
    "context"
    "go.opentelemetry.io/otel/attribute"
    "github.com/wasilak/otelgo/metrics"
)

ctx := context.Background()

config := metrics.OtelGoMetricsConfig{
    Attributes: []attribute.KeyValue{
        attribute.String("service.name", "my-service"),
    },
}

ctx, provider, err := metrics.Init(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer metrics.Shutdown(ctx, provider)
```

#### `Shutdown(ctx context.Context, meterProvider *sdk.MeterProvider) error`

Shuts down the metric provider and flushes any pending metrics. Should be called when the application is terminating to ensure all metrics are exported.

**Parameters:**
- `ctx`: The context for controlling shutdown timeout
- `meterProvider`: The provider instance to shut down

**Returns:**
- `error`: Non-nil if shutdown fails

**Example:**
```go
defer func() {
    if err := metrics.Shutdown(ctx, meterProvider); err != nil {
        log.Printf("failed to shutdown provider: %v", err)
    }
}()
```

### Metrics Builder Pattern

#### `NewBuilder() *MetricsBuilder`

Creates a new MetricsBuilder with default configuration.

**Returns:**
- `*MetricsBuilder`: A new builder instance with default configuration

#### `(*MetricsBuilder) WithAttributes(attrs ...attribute.KeyValue) *MetricsBuilder`

Sets the attributes to be added to the metric resource. This method can be chained with other builder methods.

**Parameters:**
- `attrs ...attribute.KeyValue`: The attributes to add to the metric resource

**Returns:**
- `*MetricsBuilder`: The same builder instance for method chaining

#### `(*MetricsBuilder) WithTLS(tlsConfig *internal.TLSConfig) *MetricsBuilder`

Sets the TLS configuration for the metrics exporter. Use this for custom TLS settings.

**Parameters:**
- `tlsConfig *internal.TLSConfig`: The TLS configuration to use

**Returns:**
- `*MetricsBuilder`: The same builder instance for method chaining

#### `(*MetricsBuilder) WithTLSInsecure() *MetricsBuilder`

Configures the builder to skip TLS verification (insecure mode). Use only for development/testing.

**Returns:**
- `*MetricsBuilder`: The same builder instance for method chaining

#### `(*MetricsBuilder) WithTLSCustom(insecure bool, caCertPath, clientCertPath, clientKeyPath, serverName string) *MetricsBuilder`

Configures the builder with custom TLS settings.

**Parameters:**
- `insecure bool`: Whether to skip certificate verification
- `caCertPath string`: Path to the CA certificate file
- `clientCertPath string`: Path to the client certificate file
- `clientKeyPath string`: Path to the client private key file
- `serverName string`: Override the server name for verification

**Returns:**
- `*MetricsBuilder`: The same builder instance for method chaining

#### `(*MetricsBuilder) Build(ctx context.Context) (context.Context, *sdk.MeterProvider, error)`

Builds and initializes the OpenTelemetry metric provider with the configured settings.

**Parameters:**
- `ctx context.Context`: The context for controlling initialization

**Returns:**
- `context.Context`: Updated context with metric provider
- `*sdk.MeterProvider`: Configured metric provider
- `error`: Non-nil if initialization fails

## Tracing Package

### Public Types

#### `Config`

Configuration for OpenTelemetry tracing.

**Fields:**
- `HostMetricsEnabled bool`: Specifies whether host metrics are enabled. Default is false.
- `HostMetricsInterval time.Duration`: Specifies the interval at which host metrics are collected. Default is 2 seconds.
- `RuntimeMetricsEnabled bool`: Specifies whether runtime metrics are enabled. Default is false.
- `RuntimeMetricsInterval time.Duration`: Specifies the interval at which runtime metrics are collected. Default is 2 seconds.
- `TLS *internal.TLSConfig`: TLS specifies the TLS configuration for exporters. Default is nil.

### Public Functions

#### `Init(ctx context.Context, config Config) (context.Context, *trace.TracerProvider, error)`

Initializes the OpenTelemetry tracer provider with the specified configuration. It sets up a trace pipeline by configuring exporters and resource attributes.

**Parameters:**
- `ctx`: The context for controlling tracer initialization lifetime
- `config`: The configuration containing tracer setup options and metrics settings

**Returns:**
- `context.Context`: Updated context with tracer provider
- `*trace.TracerProvider`: Configured tracer provider for creating spans
- `error`: Non-nil if initialization fails

**Example:**
```go
import (
    "context"
    "time"
    trace "go.opentelemetry.io/otel/trace"
    "github.com/wasilak/otelgo/tracing"
)

ctx := context.Background()

config := tracing.Config{
    HostMetricsEnabled: true,
    RuntimeMetricsEnabled: true,
    HostMetricsInterval: 5 * time.Second,
}

ctx, provider, err := tracing.Init(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer tracing.Shutdown(ctx, provider)
```

#### `Shutdown(ctx context.Context, tracerProvider *trace.TracerProvider) error`

Shuts down the tracer provider and flushes any pending spans. Should be called when the application is terminating to ensure all traces are exported.

**Parameters:**
- `ctx`: The context for controlling shutdown timeout
- `tracerProvider`: The provider instance to shut down

**Returns:**
- `error`: Non-nil if shutdown fails

**Example:**
```go
defer func() {
    if err := tracing.Shutdown(ctx, tracerProvider); err != nil {
        log.Printf("failed to shutdown provider: %v", err)
    }
}()
```

### Tracing Builder Pattern

#### `NewBuilder() *TracingBuilder`

Creates a new TracingBuilder with default configuration.

**Returns:**
- `*TracingBuilder`: A new builder instance with default configuration

#### `(*TracingBuilder) WithHostMetrics(enabled bool, interval time.Duration) *TracingBuilder`

Enables or disables host metrics collection with the specified interval.

**Parameters:**
- `enabled bool`: Whether to enable host metrics collection
- `interval time.Duration`: The interval between metric collections

**Returns:**
- `*TracingBuilder`: The same builder instance for method chaining

#### `(*TracingBuilder) WithRuntimeMetrics(enabled bool, interval time.Duration) *TracingBuilder`

Enables or disables runtime metrics collection with the specified interval.

**Parameters:**
- `enabled bool`: Whether to enable runtime metrics collection
- `interval time.Duration`: The interval between metric collections

**Returns:**
- `*TracingBuilder`: The same builder instance for method chaining

#### `(*TracingBuilder) WithTLS(tlsConfig *internal.TLSConfig) *TracingBuilder`

Sets the TLS configuration for the tracing exporter. Use this for custom TLS settings.

**Parameters:**
- `tlsConfig *internal.TLSConfig`: The TLS configuration to use

**Returns:**
- `*TracingBuilder`: The same builder instance for method chaining

#### `(*TracingBuilder) WithTLSInsecure() *TracingBuilder`

Configures the builder to skip TLS verification (insecure mode). Use only for development/testing.

**Returns:**
- `*TracingBuilder`: The same builder instance for method chaining

#### `(*TracingBuilder) WithTLSCustom(insecure bool, caCertPath, clientCertPath, clientKeyPath, serverName string) *TracingBuilder`

Configures the builder with custom TLS settings.

**Parameters:**
- `insecure bool`: Whether to skip certificate verification
- `caCertPath string`: Path to the CA certificate file
- `clientCertPath string`: Path to the client certificate file
- `clientKeyPath string`: Path to the client private key file
- `serverName string`: Override the server name for verification

**Returns:**
- `*TracingBuilder`: The same builder instance for method chaining

#### `(*TracingBuilder) Build(ctx context.Context) (context.Context, *trace.TracerProvider, error)`

Builds and initializes the OpenTelemetry tracer provider with the configured settings.

**Parameters:**
- `ctx context.Context`: The context for controlling initialization

**Returns:**
- `context.Context`: Updated context with tracer provider
- `*trace.TracerProvider`: Configured tracer provider
- `error`: Non-nil if initialization fails

## Slog Package

### Public Types

#### `TracingHandler`

A slog.Handler wrapper that adds OpenTelemetry tracing information to log records. Enriches log entries with trace context, span details, and other OpenTelemetry attributes.

### Public Functions

#### `NewTracingHandler(h slog.Handler) *TracingHandler`

Creates a new TracingHandler that wraps the provided slog.Handler. If the provided handler is already a TracingHandler, it returns the underlying handler to avoid multiple layers of wrapping.

**Parameters:**
- `h slog.Handler`: The slog.Handler to wrap

**Returns:**
- `*TracingHandler`: A new handler that adds tracing information to log records

**Example:**
```go
import (
    "log/slog"
    "os"
    "github.com/wasilak/otelgo/slog"
)

// Create a base handler
baseHandler := slog.NewJSONHandler(os.Stdout, nil)

// Wrap with tracing information
tracingHandler := slog.NewTracingHandler(baseHandler)
logger := slog.New(tracingHandler)

// When logging within a trace context, logs will include trace/span information
logger.Info("Processing request", slog.String("request_id", "12345"))
```

#### `(*TracingHandler) Handler() slog.Handler`

Returns the underlying slog.Handler wrapped by this TracingHandler. This method is useful when you need to access or modify the base handler.

**Returns:**
- `slog.Handler`: The underlying handler

## Internal Package

### TLS Configuration

#### `TLSConfig`

TLS configuration for OpenTelemetry exporters.

**Fields:**
- `Insecure bool`: Specifies whether TLS verification should be skipped. Default is false.
- `CACertPath string`: Path to the CA certificate file for server verification. Default is empty.
- `ClientCertPath string`: Path to the client certificate file for mutual TLS. Default is empty.
- `ClientKeyPath string`: Path to the client private key file for mutual TLS. Default is empty.
- `ServerName string`: Override the server name for TLS verification. Default is empty.

#### `NewTLSConfig() *TLSConfig`

Creates a new TLSConfig with default values based on environment variables.

**Returns:**
- `*TLSConfig`: A new TLS configuration with environment-based defaults

#### `(*TLSConfig) Validate() error`

Validates the TLS configuration to ensure it has valid settings. Checks for conflicts like using both Insecure=true and a CA certificate path.

**Returns:**
- `error`: Non-nil if the configuration is invalid

#### `(*TLSConfig) BuildTLSConfig() (*tls.Config, error)`

Builds a standard Go tls.Config from this TLSConfig. Validates the configuration and loads any required certificates.

**Returns:**
- `*tls.Config`: The built TLS configuration
- `error`: Non-nil if building the configuration fails

## Error Handling

All initialization functions return an error that should be checked:

```go
ctx, provider, err := logs.Init(ctx, config)
if err != nil {
    // Handle initialization error properly
    log.Printf("Failed to initialize logs: %v", err)
    // Application may continue without log export
    return err
}
defer logs.Shutdown(ctx, provider)  // Properly handle shutdown
```

## Environment Variables

The library respects standard OpenTelemetry environment variables:

- `OTEL_SERVICE_NAME` - Service name for resource attribution
- `OTEL_EXPORTER_OTLP_INSECURE` - Skip TLS verification (development only)
- `OTEL_EXPORTER_OTLP_CERTIFICATE` - CA certificate path
- `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` - Client certificate path
- `OTEL_EXPORTER_OTLP_CLIENT_KEY` - Client key path
- `OTEL_EXPORTER_OTLP_SERVER_NAME` - Server name override

For protocol-specific variables:

- Logs: `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL`, `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`
- Metrics: `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL`, `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` 
- Tracing: `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`, `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`

## Best Practices

1. **Always check errors** returned by Init functions and handle appropriately
2. **Always call Shutdown** when done to ensure data is properly flushed
3. **Use secure TLS by default** in production environments
4. **Provide meaningful attributes** to help with debugging and analysis
5. **Use the builder pattern** for cleaner, more readable configuration
