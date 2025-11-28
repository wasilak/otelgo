# otelgo

![GitHub tag (with filter)](https://img.shields.io/github/v/tag/wasilak/otelgo) ![GitHub go.mod Go version (branch & subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/wasilak/otelgo/main) [![Go Reference](https://pkg.go.dev/badge/github.com/wasilak/otelgo.svg)](https://pkg.go.dev/github.com/wasilak/otelgo) [![Maintainability](https://api.codeclimate.com/v1/badges/ce4cd4fe1e30b1ddbac5/maintainability)](https://codeclimate.com/github/wasilak/otelgo/maintainability)

OpenTelemetry library that unifies how Go applications implement OpenTelemetry for logs, metrics, and distributed tracing with production-ready TLS security features.

## Features

- **Logs**: OpenTelemetry log collection with configurable TLS settings
- **Metrics**: OpenTelemetry metric collection with configurable TLS settings
- **Tracing**: OpenTelemetry distributed tracing with configurable TLS settings
- **Host Metrics**: Collection of system-level metrics (CPU, memory, disk, network)
- **Runtime Metrics**: Go runtime metrics (GC, memory, goroutines, etc.)
- **Slog Integration**: Structured logging through Go's slog package
- **Builder Pattern**: Fluent APIs for easy configuration (logs.NewBuilder(), metrics.NewBuilder(), tracing.NewBuilder())
- **TLS Support**: Secure connections with comprehensive TLS configuration and validation
- **Thread Safety**: Safe for concurrent initialization and use
- **Comprehensive Testing**: Extensive test coverage (>85%) including race condition testing

## Quick Start

### Installation

```bash
go get github.com/wasilak/otelgo
```

### Basic Usage

```go
package main

import (
    "context"
    "os"

    "github.com/wasilak/otelgo/logs"
    "github.com/wasilak/otelgo/metrics"
    "github.com/wasilak/otelgo/tracing"
)

func main() {
    // Set required environment variables
    os.Setenv("OTEL_SERVICE_NAME", "my-service")

    ctx := context.Background()

    // Initialize all three providers with minimal setup
    _, logProvider, _ := logs.NewBuilder().WithTLSInsecure().Build(ctx)
    _, metricProvider, _ := metrics.NewBuilder().WithTLSInsecure().Build(ctx)
    _, traceProvider, _ := tracing.NewBuilder().WithTLSInsecure().Build(ctx)

    // Use your providers here
    // Remember to shutdown when done
    defer func() {
        _ = logs.Shutdown(ctx, logProvider)
        _ = metrics.Shutdown(ctx, metricProvider)
        _ = tracing.Shutdown(ctx, traceProvider)
    }()
}
```

## Documentation

- [Security Guide](docs/SECURITY.md) - TLS configuration, certificate setup, security best practices
- [Configuration Guide](docs/CONFIGURATION.md) - All configuration options, environment variables, and builder patterns
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Common issues, error handling, debugging techniques
- [API Reference](docs/API.md) - Complete API documentation for all public interfaces

## Configuration

### Environment Variables

| Category | Variable | Description | Default |
|----------|----------|-------------|---------|
| **General** | `OTEL_SERVICE_NAME` | Name of the service for resource attribution | (required) |
| | `OTEL_EXPORTER_OTLP_INSECURE` | Skip TLS verification (use only for development) | `false` |
| | `OTEL_EXPORTER_OTLP_CERTIFICATE` | Path to CA certificate for server verification | (none) |
| | `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` | Path to client certificate for mutual TLS | (none) |
| | `OTEL_EXPORTER_OTLP_CLIENT_KEY` | Path to client private key for mutual TLS | (none) |
| | `OTEL_EXPORTER_OTLP_SERVER_NAME` | Override server name for certificate verification | (none) |
| **Logs** | `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL` | Logs export protocol (`http` or `grpc`) | `grpc` |
| | `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | Logs export endpoint | (none) |
| **Metrics** | `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` | Metrics export protocol (`http` or `grpc`) | `grpc` |
| | `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | Metrics export endpoint | (none) |
| **Traces** | `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` | Traces export protocol (`http` or `grpc`) | `grpc` |
| | `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Traces export endpoint | (none) |

### Builder Pattern Configuration

All packages support fluent builder APIs for clean, readable configuration:

```go
// Logs with custom attributes and TLS
_, logProvider, _ := logs.NewBuilder().
    WithAttributes(attribute.String("env", "production")).
    WithTLS(&internal.TLSConfig{
        CACertPath: "/path/to/ca.pem",
        ServerName: "collector.example.com",
    }).
    Build(ctx)

// Metrics with custom attributes
_, metricProvider, _ := metrics.NewBuilder().
    WithAttributes(attribute.String("service.region", "us-west-2")).
    WithTLSInsecure(). // Development only!
    Build(ctx)

// Tracing with host/runtime metrics and TLS
_, traceProvider, _ := tracing.NewBuilder().
    WithHostMetricsEnabled(true).
    WithRuntimeMetricsEnabled(true).
    WithTLS(&internal.TLSConfig{
        ClientCertPath: "/path/to/client-cert.pem",
        ClientKeyPath:  "/path/to/client-key.pem",
    }).
    Build(ctx)
```

## Security Features

All otelgo components feature production-ready security:

- **Secure-by-default TLS**: Proper certificate validation with customizable options
- **Mutual TLS Support**: Client certificate authentication for enhanced security
- **Certificate Pinning Options**: Server name verification and custom CA certificates
- **Comprehensive Validation**: Input validation for all configuration options
- **Thread-Safe Initialization**: Safe concurrent use patterns

For detailed security configuration, see our [Security Guide](docs/SECURITY.md).

## Examples

Working examples are provided in the [examples/](examples/) directory:

- [Basic example](examples/basic/basic.go) - Fundamental usage patterns
- [TLS Custom CA example](examples/tls-custom-ca/tls-custom-ca.go) - Secure TLS configuration
- [Client Certificate example](examples/client-cert/client-cert.go) - Mutual TLS setup
- [Error Handling example](examples/error-handling/error-handling.go) - Proper error handling
- [Concurrent Initialization example](examples/concurrent-init/concurrent-init.go) - Thread-safe usage patterns

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please read our contribution guidelines and code of conduct before submitting pull requests.
