// Example: tls-custom-ca.go
//
// This example demonstrates how to configure TLS with a custom Certificate Authority (CA)
// for secure connections to OpenTelemetry collectors.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wasilak/otelgo/internal"
	"github.com/wasilak/otelgo/logs"
	"github.com/wasilak/otelgo/metrics"
	"github.com/wasilak/otelgo/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Set required environment variables for this example
	os.Setenv("OTEL_SERVICE_NAME", "tls-custom-ca-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "https://otel-collector.example.com/v1/logs")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://otel-collector.example.com/v1/metrics")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "https://otel-collector.example.com/v1/traces")

	ctx := context.Background()

	// Create custom TLS configuration with custom CA certificate
	// In a real application, these paths would point to actual certificate files
	tlsConfig := &internal.TLSConfig{
		CACertPath: "/path/to/custom-ca.crt",     // Path to your custom CA certificate
		ServerName: "otel-collector.example.com", // This should match the certificate's Common Name or Subject Alternative Name
		// Insecure: false, // This should be false in production (not set means false by default)
	}

	fmt.Println("=== TLS with Custom CA Example ===")
	fmt.Printf("Using CA certificate from: %s\n", tlsConfig.CACertPath)
	fmt.Printf("Verifying server name: %s\n", tlsConfig.ServerName)

	// Example 1: Logs with custom CA
	fmt.Println("\n--- Initializing Logs with Custom CA ---")
	logsCtx, logsProvider, logsErr := logs.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "tls-custom-ca-service"),
			attribute.String("cert.ca.type", "custom"),
		).
		WithTLS(tlsConfig).
		Build(ctx)

	if logsErr != nil {
		log.Printf("Logs provider failed to initialize (expected in example): %v", logsErr)
	} else {
		fmt.Println("Logs provider initialized successfully with custom CA")
		defer func() {
			if err := logs.Shutdown(logsCtx, logsProvider); err != nil {
				log.Printf("Error shutting down logs provider: %v", err)
			}
		}()
	}

	// Example 2: Metrics with custom CA
	fmt.Println("\n--- Initializing Metrics with Custom CA ---")
	metricsCtx, metricsProvider, metricsErr := metrics.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "tls-custom-ca-service"),
			attribute.String("cert.ca.type", "custom"),
		).
		WithTLS(tlsConfig).
		Build(ctx)

	if metricsErr != nil {
		log.Printf("Metrics provider failed to initialize (expected in example): %v", metricsErr)
	} else {
		fmt.Println("Metrics provider initialized successfully with custom CA")
		defer func() {
			if err := metrics.Shutdown(metricsCtx, metricsProvider); err != nil {
				log.Printf("Error shutting down metrics provider: %v", err)
			}
		}()
	}

	// Example 3: Tracing with custom CA
	fmt.Println("\n--- Initializing Tracing with Custom CA ---")
	tracingCtx, tracerProvider, tracingErr := tracing.NewBuilder().
		WithHostMetrics(true, 10*time.Second).
		WithRuntimeMetrics(true, 30*time.Second).
		WithTLS(tlsConfig).
		Build(ctx)

	if tracingErr != nil {
		log.Printf("Tracer provider failed to initialize (expected in example): %v", tracingErr)
	} else {
		fmt.Println("Tracer provider initialized successfully with custom CA")
		defer func() {
			if err := tracing.Shutdown(tracingCtx, tracerProvider); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()

		fmt.Println("Tracing configured with custom CA certificate")
	}

	fmt.Println("\n=== TLS Custom CA Example Completed ===")
	fmt.Println("This example demonstrates how to configure TLS with a custom Certificate Authority.")
	fmt.Println("In production, ensure:")
	fmt.Println("- CA certificate file exists and is accessible")
	fmt.Println("- Server name matches the certificate")
	fmt.Println("- Certificate is not expired")
	fmt.Println("- File permissions are secure (typically 600)")
}
