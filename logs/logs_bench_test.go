package logs

import (
	"context"
	"os"
	"testing"

	"github.com/wasilak/otelgo/internal"
	"go.opentelemetry.io/otel/attribute"
)

// BenchmarkInitBasic benchmarks the basic initialization of the logs provider with minimal configuration.
func BenchmarkInitBasic(b *testing.B) {
	// Set required environment variables for the benchmark
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http") // Use HTTP to avoid gRPC overhead in benchmark

	config := OtelGoLogsConfig{}

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

// BenchmarkInitWithAttributes benchmarks initialization with multiple attributes.
func BenchmarkInitWithAttributes(b *testing.B) {
	originalServiceName := os.Getenv("OTEL_SERVICE_NAME")
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")

	attrs := []attribute.KeyValue{
		attribute.String("service.name", "test-service"),
		attribute.String("service.version", "v1.0.0"),
		attribute.String("env", "production"),
		attribute.Int("partition", 1),
		attribute.Bool("feature.enabled", true),
	}
	config := OtelGoLogsConfig{
		Attributes: attrs,
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
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")

	config := OtelGoLogsConfig{
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
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")

	config := OtelGoLogsConfig{}

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
	originalProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	defer func() {
		os.Setenv("OTEL_SERVICE_NAME", originalServiceName)
		os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", originalProtocol)
	}()

	os.Setenv("OTEL_SERVICE_NAME", "benchmark-service")
	os.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "http")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testCtx := context.Background()
		b.StartTimer()

		_, provider, err := NewBuilder().
			WithAttributes(
				attribute.String("env", "benchmark"),
				attribute.String("iteration", "test"),
			).
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
