package common

import (
	"os"
	"strings"
)

// IsOtlpProtocolGrpc determines if the OTLP protocol is set to gRPC for a given data type.
// It checks both the specific data type environment variable and the general OTLP protocol setting.
//
// Parameters:
//   - dataType: The type of data to check (e.g., "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
//
// Returns:
//   - bool: true if the protocol is set to "grpc", false otherwise
//
// The function first checks the specific data type environment variable.
// If not set to "grpc", it falls back to checking the general "OTEL_EXPORTER_OTLP_PROTOCOL".
func IsOtlpProtocolGrpc(dataType string) bool {

	if os.Getenv(strings.ToUpper(dataType)) == "grpc" {
		return true
	}

	return os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc"
}
