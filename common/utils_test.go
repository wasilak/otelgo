package common

import (
	"os"
	"testing"
)

func TestIsOtlpProtocolGrpc(t *testing.T) {
	tests := []struct {
		name           string
		dataType       string
		envValue       string
		globalEnvValue string
		expectedResult bool
		cleanup        func()
	}{
		{
			name:           "specific protocol set to grpc",
			dataType:       "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL",
			envValue:       "grpc",
			expectedResult: true,
			cleanup: func() {
				os.Unsetenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
			},
		},
		{
			name:           "specific protocol set to http",
			dataType:       "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL",
			envValue:       "http",
			expectedResult: false,
			cleanup: func() {
				os.Unsetenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
			},
		},
		{
			name:           "specific protocol not set, global set to grpc",
			dataType:       "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL",
			globalEnvValue: "grpc",
			expectedResult: true,
			cleanup: func() {
				os.Unsetenv("OTEL_EXPORTER_OTLP_PROTOCOL")
			},
		},
		{
			name:           "specific protocol not set, global set to http",
			dataType:       "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL",
			globalEnvValue: "http",
			expectedResult: false,
			cleanup: func() {
				os.Unsetenv("OTEL_EXPORTER_OTLP_PROTOCOL")
			},
		},
		{
			name:           "neither specific nor global protocol set",
			dataType:       "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			if tt.envValue != "" {
				os.Setenv(tt.dataType, tt.envValue)
			}
			if tt.globalEnvValue != "" {
				os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", tt.globalEnvValue)
			}

			result := IsOtlpProtocolGrpc(tt.dataType)
			if result != tt.expectedResult {
				t.Errorf("IsOtlpProtocolGrpc(%q) = %v, want %v", tt.dataType, result, tt.expectedResult)
			}
		})
	}
}
