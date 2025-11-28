# Configuration Guide for otelgo

This document provides comprehensive configuration options for the otelgo library, covering all available settings, environment variables, and usage examples for various scenarios.

## Table of Contents

1. [Overview](#overview)
2. [Package Configuration Options](#package-configuration-options)
3. [Environment Variables](#environment-variables)
4. [Builder Patterns](#builder-patterns)
5. [Default Values](#default-values)
6. [Common Scenarios](#common-scenarios)

## Overview

The otelgo library provides flexible configuration options for OpenTelemetry logging, metrics, and tracing. Configuration can be specified programmatically through structs or via environment variables.

## Package Configuration Options

### Logs Package

#### OtelGoLogsConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Attributes` | `[]attribute.KeyValue` | Additional attributes to include in log records | Service name and version | 
| `TLS` | `*internal.TLSConfig` | TLS configuration for secure connections | Secure by default |

#### Example Configuration

```go
import (
    "github.com/wasilak/otelgo/logs"
    "github.com/wasilak/otelgo/internal"
    "go.opentelemetry.io/otel/attribute"
)

config := logs.OtelGoLogsConfig{
    Attributes: []attribute.KeyValue{
        attribute.String("env", "production"),
        attribute.String("service.name", "my-service"),
        attribute.String("service.version", "1.0.0"),
    },
    TLS: &internal.TLSConfig{
        CACertPath: "/path/to/ca-cert.pem",
        ServerName: "otel-collector.example.com",
    },
}
```

### Metrics Package

#### OtelGoMetricsConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Attributes` | `[]attribute.KeyValue` | Additional attributes to include in metric resource | Service name and version |
| `TLS` | `*internal.TLSConfig` | TLS configuration for secure connections | Secure by default |

#### Example Configuration

```go
import (
    "github.com/wasilak/otelgo/metrics"
    "github.com/wasilak/otelgo/internal"
    "go.opentelemetry.io/otel/attribute"
)

config := metrics.OtelGoMetricsConfig{
    Attributes: []attribute.KeyValue{
        attribute.String("region", "us-east-1"),
        attribute.String("deployment", "prod-blue"),
    },
    TLS: &internal.TLSConfig{
        Insecure: false,  // Do NOT set to true in production
    },
}
```

### Tracing Package

#### Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `HostMetricsEnabled` | `bool` | Enable collection of host metrics (CPU, memory, disk) | `false` |
| `HostMetricsInterval` | `time.Duration` | Interval between host metric collections | `2 * time.Second` |
| `RuntimeMetricsEnabled` | `bool` | Enable collection of Go runtime metrics | `false` |
| `RuntimeMetricsInterval` | `time.Duration` | Interval between runtime metric collections | `2 * time.Second` |
| `TLS` | `*internal.TLSConfig` | TLS configuration for secure connections | Secure by default |

#### Example Configuration

```go
import (
    "github.com/wasilak/otelgo/tracing"
    "github.com/wasilak/otelgo/internal"
    "time"
)

config := tracing.Config{
    HostMetricsEnabled:    true,
    HostMetricsInterval:   5 * time.Second,
    RuntimeMetricsEnabled: true,
    RuntimeMetricsInterval: 10 * time.Second,
    TLS: &internal.TLSConfig{
        ClientCertPath: "/path/to/client-cert.pem",
        ClientKeyPath:  "/path/to/client-key.pem",
    },
}
```

## Environment Variables

### Protocol Selection

| Variable | Description | Valid Values | Default |
|----------|-------------|--------------|---------|
| `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL` | Protocol for logs export | `http`/`grpc` | `grpc` |
| `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` | Protocol for metrics export | `http`/`grpc` | `grpc` |
| `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` | Protocol for traces export | `http`/`grpc` | `grpc` |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | General protocol fallback (when specific not set) | `http`/`grpc` | - |

### Endpoint Configuration

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | Logs export endpoint | `https://collector:4318/v1/logs` | - |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | Metrics export endpoint | `https://collector:4318/v1/metrics` | - |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Traces export endpoint | `https://collector:4318/v1/traces` | - |

### TLS Configuration

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `OTEL_EXPORTER_OTLP_INSECURE` | Skip certificate verification (DO NOT use in production) | `true`/`false` | `false` |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` | Path to CA certificate file | `/certs/ca.crt` | - |
| `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` | Path to client certificate file | `/certs/client.crt` | - |
| `OTEL_EXPORTER_OTLP_CLIENT_KEY` | Path to client key file | `/certs/client.key` | - |
| `OTEL_EXPORTER_OTLP_SERVER_NAME` | Override server name for certificate verification | `collector.example.com` | - |

### Service Identification

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `OTEL_SERVICE_NAME` | Name of the service for resource attribution | `my-service` | - |

## Builder Patterns

All packages provide fluent builder patterns for convenient configuration.

### Logs Builder

```go
import (
    "context"
    "github.com/wasilak/otelgo/logs"
)

ctx := context.Background()

// Basic configuration
ctx, provider, err := logs.NewBuilder().
    WithAttributes(
        attribute.String("env", "production"),
        attribute.String("version", "1.0.0"),
    ).
    WithTLSInsecure().  // Only for development
    Build(ctx)

// Production TLS configuration
ctx, provider, err = logs.NewBuilder().
    WithAttributes(
        attribute.String("env", "production"),
        attribute.Bool("feature.enabled", true),
    ).
    WithTLSCustom(false, "/path/to/ca.pem", "/path/to/cert.pem", "/path/to/key.pem", "collector.example.com").
    Build(ctx)

defer logs.Shutdown(ctx, provider)
```

### Metrics Builder

```go
import (
    "context"
    "github.com/wasilak/otelgo/metrics"
)

ctx := context.Background()

// Basic metrics configuration
ctx, provider, err := metrics.NewBuilder().
    WithAttributes(
        attribute.String("service.id", "service-123"),
        attribute.Int("partition", 1),
    ).
    WithTLSInsecure().  // Only for development
    Build(ctx)

// Production metrics configuration
ctx, provider, err = metrics.NewBuilder().
    WithAttributes(
        attribute.String("team", "backend"),
        attribute.Float64("sla", 0.999),
    ).
    WithTLS(&internal.TLSConfig{
        CACertPath: "/secure/path/ca-cert.pem",
        ServerName: "metrics-collector.example.com",
    }).
    Build(ctx)

defer metrics.Shutdown(ctx, provider)
```

### Tracing Builder

```go
import (
    "context"
    "github.com/wasilak/otelgo/tracing"
    "time"
)

ctx := context.Background()

// Basic tracing configuration
ctx, provider, err := tracing.NewBuilder().
    WithHostMetrics(true, 5*time.Second).
    WithRuntimeMetrics(true, 10*time.Second).
    WithTLS(&internal.TLSConfig{
        ClientCertPath: "/client/cert.pem",
        ClientKeyPath:  "/client/key.pem",
    }).
    Build(ctx)

// Minimal tracing configuration
ctx, provider, err = tracing.NewBuilder().
    WithTLSInsecure().  // Only for development
    Build(ctx)

defer tracing.Shutdown(ctx, provider)
```

## Default Values

When specific configuration is not provided, the following defaults are applied:

### Global Defaults

| Config Field | Default Value | Notes |
|--------------|---------------|-------|
| Service Name | Value of `OTEL_SERVICE_NAME` environment variable | Required for proper attribution |
| Service Version | `v0.0.0` | Can be overridden |
| TLS Insecure | `false` | Secure connections enabled by default |
| TLS CA Cert Path | Empty | Verify against system trust store |

### Package-Specific Defaults

#### Logs Package
- Attributes: `service.name` from `OTEL_SERVICE_NAME`, `service.version` as `v0.0.0`

#### Metrics Package  
- Attributes: `service.name` from `OTEL_SERVICE_NAME`, `service.version` as `v0.0.0`

#### Tracing Package
- `HostMetricsEnabled`: `false`
- `HostMetricsInterval`: `2 * time.Second`
- `RuntimeMetricsEnabled`: `false` 
- `RuntimeMetricsInterval`: `2 * time.Second`
- `Attributes`: `service.name` from `OTEL_SERVICE_NAME`, `service.version` as `v0.0.0`

## Common Scenarios

### Development Setup (Insecure)

```go
import (
    "context"
    "os"
    "github.com/wasilak/otelgo/logs"
    "github.com/wasilak/otelgo/metrics"
    "github.com/wasilak/otelgo/tracing"
)

// Set environment variables for development
os.Setenv("OTEL_SERVICE_NAME", "dev-service")
os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")

// Initialize all providers with insecure TLS
ctx := context.Background()

// Logs
ctx, logProvider, _ := logs.NewBuilder().
    WithTLSInsecure().
    Build(ctx)
defer logs.Shutdown(ctx, logProvider)

// Metrics  
ctx, metricProvider, _ := metrics.NewBuilder().
    WithTLSInsecure().
    Build(ctx)
defer metrics.Shutdown(ctx, metricProvider)

// Tracing
ctx, traceProvider, _ := tracing.NewBuilder().
    WithTLSInsecure().
    Build(ctx)
defer tracing.Shutdown(ctx, traceProvider)
```

### Production Setup (Secure)

```go
import (
    "context"
    "time"
    "github.com/wasilak/otelgo/tracing"
)

// Production tracing with metrics and secure TLS
ctx := context.Background()

ctx, provider, err := tracing.NewBuilder().
    WithHostMetrics(true, 30*time.Second).      // Monitor host resources
    WithRuntimeMetrics(true, 15*time.Second).   // Monitor Go runtime
    WithTLS(&internal.TLSConfig{
        CACertPath:     "/secure/certs/ca.pem",
        ClientCertPath: "/secure/certs/client.pem",
        ClientKeyPath:  "/secure/certs/client-key.pem",
        ServerName:     "prod-collector.example.com",
    }).
    WithAttributes(
        attribute.String("environment", "production"),
        attribute.String("region", "us-west-2"),
        attribute.Bool("monitoring.enabled", true),
    ).
    Build(ctx)

if err != nil {
    panic(err)
}

defer tracing.Shutdown(ctx, provider)
```

### Environment-Based Configuration

```go
import (
    "os"
    "github.com/wasilak/otelgo/logs"
    "github.com/wasilak/otelgo/internal"
)

// Environment-specific configuration
env := os.Getenv("ENVIRONMENT")

var tlsConfig *internal.TLSConfig
switch env {
case "development":
    tlsConfig = &internal.TLSConfig{Insecure: true}
case "staging":
    tlsConfig = &internal.TLSConfig{
        CACertPath: "/staging/certs/ca.pem",
        ServerName: "staging-collector.example.com",
    }
case "production":
    tlsConfig = &internal.TLSConfig{
        CACertPath:     "/prod/certs/ca.pem",
        ClientCertPath: "/prod/certs/client.pem",
        ClientKeyPath:  "/prod/certs/client-key.pem",
        ServerName:     "collector.corp.internal",
    }
default:
    tlsConfig = nil  // Use system defaults
}

config := logs.OtelGoLogsConfig{TLS: tlsConfig}
```

## Advanced Configuration Examples

### Custom Attribute Processing

```go
import (
    "go.opentelemetry.io/otel/attribute"
    "github.com/wasilak/otelgo/logs"
)

// Combine multiple sources of attributes
attrs := []attribute.KeyValue{
    attribute.String("service.name", os.Getenv("SERVICE_NAME")),
    attribute.String("instance.id", os.Getenv("INSTANCE_ID")),
    attribute.Bool("debug", os.Getenv("DEBUG_MODE") == "true"),
}

// Add environment-specific attributes
if env := os.Getenv("ENVIRONMENT"); env != "" {
    attrs = append(attrs, attribute.String("environment", env))
}

config := logs.OtelGoLogsConfig{
    Attributes: attrs,
    TLS: &internal.TLSConfig{
        ServerName: os.Getenv("OTEL_SERVER_NAME"),
    },
}
```

## Migration Guide

### From Basic Configuration to Builder Pattern

**Before (Basic Config):**
```go
// Old way
config := logs.OtelGoLogsConfig{
    Attributes: []attribute.KeyValue{
        // ... attributes
    },
    TLS: &internal.TLSConfig{
        // ... TLS config
    },
}
ctx, provider, err := logs.Init(ctx, config)
```

**After (Builder Pattern):**
```go
// New way - more readable and extensible
ctx, provider, err := logs.NewBuilder().
    WithAttributes(
        // ... attributes
    ).
    WithTLS(&internal.TLSConfig{
        // ... TLS config
    }).
    Build(ctx)
```

---

*This guide was last updated for otelgo version 1.0.0. Always refer to the latest documentation for the most current configuration options and best practices.*