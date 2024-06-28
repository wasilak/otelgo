package common

import (
	"os"
	"strings"
)

func CheckOtlpProtocol(dataType string, protocol string) bool {

	protocolLower := strings.ToLower(protocol)

	switch strings.ToLower(dataType) {
	case "traces":
		if os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL") == protocolLower {
			return true
		}
	case "logs":
		if os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL") == protocolLower {
			return true
		}
	case "metrics":
		if os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL") == protocolLower {
			return true
		}
	}

	return os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == protocolLower
}
