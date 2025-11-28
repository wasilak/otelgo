// Example: concurrent-init.go
//
// This example demonstrates safe concurrent initialization of multiple otelgo providers.
// It shows how to initialize logs, metrics, and tracing from multiple goroutines safely.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/wasilak/otelgo/logs"
	"github.com/wasilak/otelgo/metrics"
	"github.com/wasilak/otelgo/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Set environment variables for this example
	os.Setenv("OTEL_SERVICE_NAME", "concurrent-init-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	ctx := context.Background()

	fmt.Println("=== Concurrent Initialization Example ===")
	fmt.Println("Demonstrating safe concurrent initialization of providers")

	// Example 1: Concurrent initialization of same provider type from multiple goroutines
	fmt.Println("--- Concurrent Same-Type Provider Initialization ---")
	concurrentSameTypeExample(ctx)

	// Example 2: Concurrent initialization of different provider types
	fmt.Println("\n--- Concurrent Cross-Provider Initialization ---")
	concurrentCrossProviderExample(ctx)

	// Example 3: Mixed usage pattern
	fmt.Println("\n--- Mixed Concurrent Usage Pattern ---")
	mixedConcurrentExample(ctx)

	fmt.Println("\n=== Concurrent Initialization Example Completed ===")
	fmt.Println("Safety notes:")
	fmt.Println("- Each goroutine should use its own provider instance")
	fmt.Println("- Providers are generally safe for concurrent use by design")
	fmt.Println("- Always ensure proper shutdown of each provider")
	fmt.Println("- Resource isolation between goroutines is maintained")
}

// concurrentSameTypeExample shows initializing the same type of provider from multiple goroutines
func concurrentSameTypeExample(ctx context.Context) {
	const numGoroutines = 10
	var wg sync.WaitGroup

	fmt.Printf("Initializing %d log providers concurrently...\n", numGoroutines)

	// Use a mutex to prevent console output intermixing (not needed for safety, just for readability)
	var consoleMutex sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine gets its own provider instance
			goroutineCtx, provider, err := logs.NewBuilder().
				WithAttributes(
					attribute.String("service.name", "concurrent-logs-service"),
					attribute.String("goroutine.id", fmt.Sprintf("%d", id)),
					attribute.String("init.type", "concurrent"),
				).
				WithTLSInsecure(). // Only for development
				Build(ctx)

			if err != nil {
				consoleMutex.Lock()
				log.Printf("Goroutine %d: Failed to initialize log provider: %v", id, err)
				consoleMutex.Unlock()
				return
			}

			// Simulate some work with the provider
			time.Sleep(10 * time.Millisecond) // Simulate real usage

			// Each goroutine must shut down its own provider
			if err := logs.Shutdown(goroutineCtx, provider); err != nil {
				consoleMutex.Lock()
				log.Printf("Goroutine %d: Error during shutdown: %v", id, err)
				consoleMutex.Unlock()
			}

			consoleMutex.Lock()
			fmt.Printf("  - Goroutine %d: Log provider initialized and shut down\n", id)
			consoleMutex.Unlock()
		}(i)
	}

	wg.Wait()
	fmt.Println("  - All log providers handled concurrently")
}

// concurrentCrossProviderExample shows initializing different provider types concurrently
func concurrentCrossProviderExample(ctx context.Context) {
	var wg sync.WaitGroup
	results := make(chan string, 3) // Channel to collect results

	// Goroutine 1: Initialize Logs
	wg.Add(1)
	go func() {
		defer wg.Done()

		logsCtx, provider, err := logs.NewBuilder().
			WithAttributes(
				attribute.String("service.name", "cross-provider-service"),
				attribute.String("component", "logging"),
				attribute.String("concurrent.init", "true"),
			).
			WithTLSInsecure().
			Build(ctx)

		if err != nil {
			results <- fmt.Sprintf("Logs failed: %v", err)
			return
		}

		time.Sleep(5 * time.Millisecond) // Simulate usage

		if err := logs.Shutdown(logsCtx, provider); err != nil {
			results <- fmt.Sprintf("Logs shutdown failed: %v", err)
			return
		}

		results <- "Logs initialized and shut down successfully"
	}()

	// Goroutine 2: Initialize Metrics
	wg.Add(1)
	go func() {
		defer wg.Done()

		metricsCtx, provider, err := metrics.NewBuilder().
			WithAttributes(
				attribute.String("service.name", "cross-provider-service"),
				attribute.String("component", "metrics"),
				attribute.String("concurrent.init", "true"),
			).
			WithTLSInsecure().
			Build(ctx)

		if err != nil {
			results <- fmt.Sprintf("Metrics failed: %v", err)
			return
		}

		time.Sleep(5 * time.Millisecond) // Simulate usage

		if err := metrics.Shutdown(metricsCtx, provider); err != nil {
			results <- fmt.Sprintf("Metrics shutdown failed: %v", err)
			return
		}

		results <- "Metrics initialized and shut down successfully"
	}()

	// Goroutine 3: Initialize Tracing
	wg.Add(1)
	go func() {
		defer wg.Done()

		tracingCtx, provider, err := tracing.NewBuilder().
			WithHostMetrics(false, 2*time.Second).    // Disable for example
			WithRuntimeMetrics(false, 2*time.Second). // Disable for example
			WithTLSInsecure().
			Build(ctx)

		if err != nil {
			results <- fmt.Sprintf("Tracing failed: %v", err)
			return
		}

		time.Sleep(5 * time.Millisecond) // Simulate usage

		if err := tracing.Shutdown(tracingCtx, provider); err != nil {
			results <- fmt.Sprintf("Tracing shutdown failed: %v", err)
			return
		}

		results <- "Tracing initialized and shut down successfully"
	}()

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Print results
	for result := range results {
		fmt.Printf("  - %s\n", result)
	}
}

// mixedConcurrentExample demonstrates a more complex usage scenario
func mixedConcurrentExample(ctx context.Context) {
	var wg sync.WaitGroup

	// Create multiple services each with their own providers
	services := []string{"auth-service", "payment-service", "notification-service"}

	for i, serviceName := range services {
		wg.Add(1)
		go func(serviceIndex int, name string) {
			defer wg.Done()

			// Each service gets its own context and provider set
			serviceCtx := ctx

			// Initialize all three providers for this service
			logsCtx, logsProvider, logsErr := logs.NewBuilder().
				WithAttributes(
					attribute.String("service.name", name),
					attribute.String("service.group", "mixed-concurrent"),
					attribute.Int("service.index", serviceIndex),
				).
				WithTLSInsecure().
				Build(serviceCtx)

			if logsErr == nil {
				defer logs.Shutdown(logsCtx, logsProvider)
			}

			metricsCtx, metricsProvider, metricsErr := metrics.NewBuilder().
				WithAttributes(
					attribute.String("service.name", name),
					attribute.String("service.group", "mixed-concurrent"),
					attribute.Int("service.index", serviceIndex),
				).
				WithTLSInsecure().
				Build(serviceCtx)

			if metricsErr == nil {
				defer metrics.Shutdown(metricsCtx, metricsProvider)
			}

			tracingSvcCtx, tracingProvider, tracingErr := tracing.NewBuilder().
				WithHostMetrics(false, 2*time.Second).
				WithRuntimeMetrics(false, 2*time.Second).
				WithTLSInsecure().
				Build(serviceCtx)

			if tracingErr == nil {
				defer tracing.Shutdown(tracingSvcCtx, tracingProvider)
			}

			// Simulate service activity
			time.Sleep(15 * time.Millisecond)

			// Report on this service's initialization
			errors := []string{}
			if logsErr != nil {
				errors = append(errors, "logs")
			}
			if metricsErr != nil {
				errors = append(errors, "metrics")
			}
			if tracingErr != nil {
				errors = append(errors, "tracing")
			}

			if len(errors) > 0 {
				fmt.Printf("  - Service %s had initialization errors for: %v\n", serviceName, errors)
			} else {
				fmt.Printf("  - Service %s fully initialized all providers\n", serviceName)
			}
		}(i, serviceName)
	}

	wg.Wait()
}
