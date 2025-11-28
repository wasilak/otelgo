package tracing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/wasilak/otelgo/internal"
)

// BenchmarkInitBasic benchmarks the basic initialization of the tracing provider with minimal configuration.
func BenchmarkInitBasic(b *testing.B) {
	// Set required environment variables for the benchmark
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http") // Use HTTP to avoid gRPC overhead in benchmark

	config := Config{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Clear context between iterations if needed
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := Init(testCtx, config)
		if err != nil {
			b.Fatalf("Init failed: %v", err)
		}

		b.StopTimer()
		if provider != nil {
			_ = Shutdown(testCtx, provider)
		}
		b.StartTimer()
	}
}

// BenchmarkInitWithHostMetrics benchmarks initialization with host metrics enabled.
func BenchmarkInitWithHostMetrics(b *testing.B) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	config := Config{
		HostMetricsEnabled:    true,
		HostMetricsInterval:   10 * time.Second,
		RuntimeMetricsEnabled: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := Init(testCtx, config)
		if err != nil {
			b.Fatalf("Init failed: %v", err)
		}

		b.StopTimer()
		if provider != nil {
			_ = Shutdown(testCtx, provider)
		}
		b.StartTimer()
	}
}

// BenchmarkInitWithRuntimeMetrics benchmarks initialization with runtime metrics enabled.
func BenchmarkInitWithRuntimeMetrics(b *testing.B) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	config := Config{
		HostMetricsEnabled:     false,
		RuntimeMetricsEnabled:  true,
		RuntimeMetricsInterval: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := Init(testCtx, config)
		if err != nil {
			b.Fatalf("Init failed: %v", err)
		}

		b.StopTimer()
		if provider != nil {
			_ = Shutdown(testCtx, provider)
		}
		b.StartTimer()
	}
}

// BenchmarkInitWithBothMetrics benchmarks initialization with both host and runtime metrics enabled.
func BenchmarkInitWithBothMetrics(b *testing.B) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	config := Config{
		HostMetricsEnabled:     true,
		HostMetricsInterval:    15 * time.Second,
		RuntimeMetricsEnabled:  true,
		RuntimeMetricsInterval: 7 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := Init(testCtx, config)
		if err != nil {
			b.Fatalf("Init failed: %v", err)
		}

		b.StopTimer()
		if provider != nil {
			_ = Shutdown(testCtx, provider)
		}
		b.StartTimer()
	}
}

// BenchmarkInitWithTLS benchmarks initialization with TLS configuration.
func BenchmarkInitWithTLS(b *testing.B) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	config := Config{
		TLS: &internal.TLSConfig{
			Insecure: true, // Use insecure for benchmark to avoid file I/O
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := Init(testCtx, config)
		if err != nil {
			b.Fatalf("Init failed: %v", err)
		}

		b.StopTimer()
		if provider != nil {
			_ = Shutdown(testCtx, provider)
		}
		b.StartTimer()
	}
}

// BenchmarkInitMemory benchmarks the memory allocation during initialization.
func BenchmarkInitMemory(b *testing.B) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	config := Config{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := Init(testCtx, config)
		if err != nil {
			b.Fatalf("Init failed: %v", err)
		}

		b.StopTimer()
		if provider != nil {
			_ = Shutdown(testCtx, provider)
		}
		b.StartTimer()
	}
}

// BenchmarkBuilderPattern benchmarks the builder pattern initialization.
func BenchmarkBuilderPattern(b *testing.B) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := NewBuilder().
			WithHostMetrics(true, 5*time.Second).
			WithRuntimeMetrics(true, 5*time.Second).
			WithTLSInsecure().
			Build(testCtx)
		if err != nil {
			b.Fatalf("Builder Build failed: %v", err)
		}

		b.StopTimer()
		if provider != nil {
			_ = Shutdown(testCtx, provider)
		}
		b.StartTimer()
	}
}