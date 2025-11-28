// Example: client-cert.go
//
// This example demonstrates how to configure mutual TLS (mTLS) with client certificates
// for authenticating with OpenTelemetry collectors.
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
	os.Setenv("OTEL_SERVICE_NAME", "mtls-client-cert-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "https://secure-otel-collector.example.com/v1/logs")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://secure-otel-collector.example.com/v1/metrics")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "https://secure-otel-collector.example.com/v1/traces")

	ctx := context.Background()

	// Create TLS configuration for mutual TLS
	// In a real application, these paths would point to actual certificate files
	mTLSTLSConfig := &internal.TLSConfig{
		Insecure:       false,                               // Never insecure for mTLS
		CACertPath:     "/path/to/ca-cert.pem",              // CA that issued the server certificate
		ClientCertPath: "/path/to/client-cert.pem",          // Client certificate
		ClientKeyPath:  "/path/to/client-key.pem",           // Client private key
		ServerName:     "secure-otel-collector.example.com", // Server name for SNI and cert validation
	}

	fmt.Println("=== Mutual TLS (mTLS) Client Certificate Example ===")
	fmt.Printf("Using client certificate: %s\n", mTLSTLSConfig.ClientCertPath)
	fmt.Printf("Using client key: %s\n", mTLSTLSConfig.ClientKeyPath)
	fmt.Printf("Verifying against CA: %s\n", mTLSTLSConfig.CACertPath)
	fmt.Printf("Server name verification: %s\n", mTLSTLSConfig.ServerName)

	// Example 1: Logs with client certificate
	fmt.Println("\n--- Initializing Logs with Client Certificate ---")
	logsCtx, logsProvider, logsErr := logs.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "mtls-logs-service"),
			attribute.String("auth.method", "client-certificate"),
			attribute.String("cert.purpose", "authentication"),
		).
		WithTLS(mTLSTLSConfig).
		Build(ctx)

	if logsErr != nil {
		log.Printf("Warning: Logs provider failed to initialize (expected in example): %v", logsErr)
	} else {
		fmt.Println("Logs provider initialized successfully with client certificate")
		defer func() {
			if err := logs.Shutdown(logsCtx, logsProvider); err != nil {
				log.Printf("Error shutting down logs provider: %v", err)
			}
		}()
	}

	// Example 2: Metrics with client certificate
	fmt.Println("\n--- Initializing Metrics with Client Certificate ---")
	metricsCtx, metricsProvider, metricsErr := metrics.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "mtls-metrics-service"),
			attribute.String("auth.method", "mutual-tls"),
			attribute.String("cert.type", "client"),
		).
		WithTLS(mTLSTLSConfig).
		Build(ctx)

	if metricsErr != nil {
		log.Printf("Warning: Metrics provider failed to initialize (expected in example): %v", metricsErr)
	} else {
		fmt.Println("Metrics provider initialized successfully with client certificate")
		defer func() {
			if err := metrics.Shutdown(metricsCtx, metricsProvider); err != nil {
				log.Printf("Error shutting down metrics provider: %v", err)
			}
		}()
	}

	// Example 3: Tracing with client certificate
	fmt.Println("\n--- Initializing Tracing with Client Certificate ---")
	tracingCtx, tracerProvider, tracingErr := tracing.NewBuilder().
		WithHostMetrics(true, 10*time.Second).
		WithRuntimeMetrics(true, 5*time.Second).
		WithTLS(mTLSTLSConfig).
		Build(ctx)

	if tracingErr != nil {
		log.Printf("Warning: Tracer provider failed to initialize (expected in example): %v", tracingErr)
	} else {
		fmt.Println("Tracer provider initialized successfully with client certificate")
		defer func() {
			if err := tracing.Shutdown(tracingCtx, tracerProvider); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()

		fmt.Println("All providers configured with mutual TLS authentication")
	}

	fmt.Println("\n=== Client Certificate (mTLS) Example Completed ===")
	fmt.Println("This example demonstrates mutual TLS setup with client certificates.")
	fmt.Println("In a production environment, ensure:")
	fmt.Println("- Client certificate and key files exist and are accessible")
	fmt.Println("- CA certificate is trusted")
	fmt.Println("- File permissions are secure (especially the private key)")
	fmt.Println("- Certificate is not expired")
	fmt.Println("- Server name matches the certificate")
	fmt.Println("\nSecurity Notes:")
	fmt.Println("- Store private keys securely with appropriate permissions (600)")
	fmt.Println("- Use short-lived certificates if possible")
	fmt.Println("- Implement certificate rotation procedures")
	fmt.Println("- Monitor certificate expiration dates")
}
