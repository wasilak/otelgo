package common

import (
	"os"
	"strings"
)

func IsOtlpProtocolGrpc(dataType string) bool {

	if os.Getenv(strings.ToUpper(dataType)) == "grpc" {
		return true
	}

	return os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "grpc"
}
