// Example: error-handling.go
//
// This example demonstrates proper error handling when initializing and using
// the otelgo library components. It shows how to handle initialization errors,
// connection failures, and graceful degradation strategies.
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
	// Set environment for this example
	os.Setenv("OTEL_SERVICE_NAME", "error-handling-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	ctx := context.Background()

	fmt.Println("=== Error Handling Example ===")
	fmt.Println("This example demonstrates proper error handling strategies")

	// Example 1: Graceful handling of initialization errors
	fmt.Println("\n--- Safe Initialization with Error Handling ---")

	// Attempt 1: With potentially problematic TLS config (for demonstration)
	tlsConfigWithErrors := &internal.TLSConfig{
		CACertPath:     "/nonexistent/ca-cert.pem", // This will cause an error
		ClientCertPath: "/nonexistent/client-cert.pem",
		ClientKeyPath:  "/nonexistent/client-key.pem",
		ServerName:     "collector.example.com",
	}

	logsCtx, logsProvider, logsErr := logs.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "error-handling-service"),
			attribute.String("scenario", "tls-error-demo"),
		).
		WithTLS(tlsConfigWithErrors). // This will fail
		Build(ctx)

	if logsErr != nil {
		// Properly handle the initialization error
		log.Printf("INFO: Logs provider failed to initialize (expected): %v", logsErr)
		log.Println("  -> Application can continue without log export (fallback to console or disable)")
		log.Println("  -> This is normal behavior when collector is unavailable")
	} else {
		fmt.Println("Logs provider initialized successfully")
		defer func() {
			if err := logs.Shutdown(logsCtx, logsProvider); err != nil {
				log.Printf("Error during logs provider shutdown: %v", err)
			}
		}()
	}

	// Example 2: Recovery from error with fallback configuration
	fmt.Println("\n--- Fallback Configuration Strategy ---")
	// If the secure config fails, try with insecure for development
	fallbackCtx, fallbackProvider, fallbackErr := logs.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "error-handling-service"),
			attribute.String("strategy", "fallback"),
		).
		WithTLSInsecure(). // Fallback to insecure for demo purposes only!
		Build(ctx)

	if fallbackErr != nil {
		log.Printf("FATAL: All initialization attempts failed: %v", fallbackErr)
		return // In a real app, you might have other fallbacks or continue without telemetry
	} else {
		fmt.Println("Fallback provider initialized successfully")
		defer func() {
			if err := logs.Shutdown(fallbackCtx, fallbackProvider); err != nil {
				log.Printf("Error during fallback provider shutdown: %v", err)
			}
		}()
	}

	// Example 3: Metrics with error handling
	fmt.Println("\n--- Safe Metrics Initialization ---")
	metricsCtx, metricsProvider, metricsErr := metrics.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "metrics-error-demo"),
			attribute.String("purpose", "error-handling"),
		).
		WithTLS(&internal.TLSConfig{
			// Example of misconfiguration that causes error
			Insecure:   true,
			CACertPath: "/some/path", // Can't have both Insecure=true and CA cert
		}).
		Build(ctx)

	if metricsErr != nil {
		log.Printf("WARN: Metrics provider failed (expected due to conflicting TLS config): %v", metricsErr)
		log.Println("  -> Continuing without metrics export")
	} else {
		fmt.Println("Metrics provider initialized successfully")
		defer func() {
			if err := metrics.Shutdown(metricsCtx, metricsProvider); err != nil {
				log.Printf("Error during metrics provider shutdown: %v", err)
			}
		}()
	}

	// Example 4: Safe tracing initialization
	fmt.Println("\n--- Safe Tracing Initialization ---")
	tracingCtx, tracerProvider, tracingErr := tracing.NewBuilder().
		WithHostMetrics(true, 10*time.Second).
		WithRuntimeMetrics(true, 15*time.Second).
		WithTLS(&internal.TLSConfig{
			CACertPath: "/nonexistent/ca.pem", // Will cause error
		}).
		Build(ctx)

	if tracingErr != nil {
		log.Printf("INFO: Tracer provider failed to initialize (expected): %v", tracingErr)
		log.Println("  -> Application continues without distributed tracing")
	} else {
		fmt.Println("Tracer provider initialized successfully")
		defer func() {
			if err := tracing.Shutdown(tracingCtx, tracerProvider); err != nil {
				log.Printf("Error during tracer provider shutdown: %v", err)
			}
		}()

		// Example of runtime error handling during tracing usage
		fmt.Println("\n--- Runtime Error Handling during Tracing ---")
		// Even if initialization succeeds, network errors can occur later
		// The provider handles this internally, but we should still handle shutdown errors
	}

	// Example 5: Comprehensive error handling strategy
	fmt.Println("\n--- Comprehensive Error Handling Pattern ---")

	// Initialize multiple providers with individual error handling
	initializeAllProviders(ctx)

	fmt.Println("\n=== Error Handling Example Completed ===")
	fmt.Println("Key takeaways:")
	fmt.Println("1. Always check initialization errors")
	fmt.Println("2. Implement fallback configurations when appropriate")
	fmt.Println("3. Properly handle shutdown operations")
	fmt.Println("4. Log errors appropriately without crashing the application")
	fmt.Println("5. Allow the application to continue even if telemetry fails")
}

// initializeAllProviders demonstrates a pattern for initializing multiple providers
// with appropriate error handling
func initializeAllProviders(ctx context.Context) {
	// Initialize logs
	logsCtx, logsProvider, logsErr := logs.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "multi-provider-demo"),
			attribute.String("init.phase", "error-handling"),
		).
		WithTLSInsecure(). // Only for development/debugging
		Build(ctx)

	if logsErr != nil {
		log.Printf("WARNING: Logs provider failed to initialize: %v", logsErr)
	} else {
		defer func() {
			if err := logs.Shutdown(logsCtx, logsProvider); err != nil {
				log.Printf("Issue during logs shutdown: %v", err) // Use log.Printf, not fatal
			}
		}()
		fmt.Println("  - Logs initialized successfully")
	}

	// Initialize metrics
	metricsCtx, metricsProvider, metricsErr := metrics.NewBuilder().
		WithAttributes(
			attribute.String("service.name", "multi-provider-demo"),
			attribute.String("init.phase", "error-handling"),
		).
		WithTLSInsecure(). // Only for development/debugging
		Build(ctx)

	if metricsErr != nil {
		log.Printf("WARNING: Metrics provider failed to initialize: %v", metricsErr)
	} else {
		defer func() {
			if err := metrics.Shutdown(metricsCtx, metricsProvider); err != nil {
				log.Printf("Issue during metrics shutdown: %v", err)
			}
		}()
		fmt.Println("  - Metrics initialized successfully")
	}

	// Initialize tracing
	tracingCtx, tracerProvider, tracingErr := tracing.NewBuilder().
		WithHostMetrics(true, 10*time.Second).
		WithRuntimeMetrics(true, 15*time.Second).
		WithTLSInsecure(). // Only for development/debugging
		Build(ctx)

	if tracingErr != nil {
		log.Printf("INFO: Tracer provider failed to initialize: %v", tracingErr)
	} else {
		defer func() {
			if err := tracing.Shutdown(tracingCtx, tracerProvider); err != nil {
				log.Printf("Issue during tracing shutdown: %v", err)
			}
		}()
		fmt.Println("  - Tracing initialized successfully")
	}

	fmt.Println("  - All providers handled with appropriate error handling")
}
