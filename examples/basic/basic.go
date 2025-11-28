// Example: basic.go
//
// This example demonstrates basic usage of the otelgo library for logs, metrics, and tracing.
// It shows simple initialization and usage patterns for all three components.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wasilak/otelgo/logs"
	"github.com/wasilak/otelgo/metrics"
	"github.com/wasilak/otelgo/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Set required environment variables for this example
	_ = os.Setenv("OTEL_SERVICE_NAME", "basic-example-service")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")
	// Note: These endpoints won't work in a real scenario without a collector
	// For this example, they're set to illustrate the configuration
	_ = os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://localhost:4318/v1/logs")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4318/v1/metrics")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://localhost:4318/v1/traces")

	ctx := context.Background()

	// Example 1: Basic Logs Usage
	fmt.Println("=== Basic Logs Example ===")
	logsCtx, logsProvider, logsErr := logs.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "basic-example-service"),
			attribute.String("example.type", "basic"),
		).
		WithTLSInsecure(). // Only for development
		Build(ctx)

	if logsErr != nil {
		log.Printf("Warning: Could not initialize logs provider: %v", logsErr)
	} else {
		fmt.Println("Logs provider initialized successfully")
		defer func() {
			if err := logs.Shutdown(logsCtx, logsProvider); err != nil {
				log.Printf("Error shutting down logs provider: %v", err)
			}
		}()
	}

	// Example 2: Basic Metrics Usage
	fmt.Println("\n=== Basic Metrics Example ===")
	metricsCtx, metricsProvider, metricsErr := metrics.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "basic-example-service"),
			attribute.String("example.type", "basic"),
		).
		WithTLSInsecure(). // Only for development
		Build(ctx)

	if metricsErr != nil {
		log.Printf("Warning: Could not initialize metrics provider: %v", metricsErr)
	} else {
		fmt.Println("Metrics provider initialized successfully")
		defer func() {
			if err := metrics.Shutdown(metricsCtx, metricsProvider); err != nil {
				log.Printf("Error shutting down metrics provider: %v", err)
			}
		}()
	}

	// Example 3: Basic Tracing Usage
	fmt.Println("\n=== Basic Tracing Example ===")
	tracingCtx, tracerProvider, tracingErr := tracing.NewBuilder().
		WithHostMetrics(false, 2*time.Second).
		WithRuntimeMetrics(false, 2*time.Second).
		WithTLSInsecure(). // Only for development
		Build(ctx)

	if tracingErr != nil {
		log.Printf("Warning: Could not initialize tracer provider: %v", tracingErr)
	} else {
		fmt.Println("Tracer provider initialized successfully")
		defer func() {
			if err := tracing.Shutdown(tracingCtx, tracerProvider); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()

		fmt.Println("Tracing initialized successfully")
	}

	fmt.Println("\n=== Basic Example Completed ===")
	fmt.Println("This example demonstrates initialization of logs, metrics, and tracing providers")
	fmt.Println("with basic attributes and insecure TLS (for development only)")
}
