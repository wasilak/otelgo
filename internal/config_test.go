package internal

import (
	"testing"
	"time"
)

func TestConfigValidator_ValidateEndpoint(t *testing.T) {
	cv := NewConfigValidator()

	tests := []struct {
		name      string
		endpoint  string
		expectErr bool
	}{
		{
			name:      "valid http endpoint",
			endpoint:  "http://localhost:4318",
			expectErr: false,
		},
		{
			name:      "valid https endpoint",
			endpoint:  "https://example.com:443",
			expectErr: false,
		},
		{
			name:      "valid grpc endpoint",
			endpoint:  "grpc://localhost:4317",
			expectErr: false,
		},
		{
			name:      "valid grpcs endpoint",
			endpoint:  "grpcs://localhost:4317",
			expectErr: false,
		},
		{
			name:      "empty endpoint",
			endpoint:  "",
			expectErr: true,
		},
		{
			name:      "invalid URL format",
			endpoint:  "not a url",
			expectErr: true,
		},
		{
			name:      "unsupported scheme",
			endpoint:  "ftp://example.com",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateEndpoint(tt.endpoint)

			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateEndpoint() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestConfigValidator_ValidateInterval(t *testing.T) {
	cv := NewConfigValidator()

	tests := []struct {
		name      string
		interval  time.Duration
		expectErr bool
	}{
		{
			name:      "valid positive interval",
			interval:  10 * time.Second,
			expectErr: false,
		},
		{
			name:      "zero interval",
			interval:  0,
			expectErr: true,
		},
		{
			name:      "negative interval",
			interval:  -5 * time.Second,
			expectErr: true,
		},
		{
			name:      "very large interval",
			interval:  25 * time.Hour,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateInterval(tt.interval)

			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateInterval() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestConfigValidator_ValidateTLSConfig(t *testing.T) {
	cv := NewConfigValidator()

	tests := []struct {
		name      string
		tlsConfig *TLSConfig
		expectErr bool
	}{
		{
			name:      "nil TLS config",
			tlsConfig: nil,
			expectErr: true,
		},
		{
			name: "valid insecure TLS config",
			tlsConfig: &TLSConfig{
				Insecure: true,
			},
			expectErr: false,
		},
		{
			name: "conflicting TLS config",
			tlsConfig: &TLSConfig{
				Insecure:   true,
				CACertPath: "/path/to/cert.pem",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateTLSConfig(tt.tlsConfig)

			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateTLSConfig() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestConfigValidator_ValidateServiceName(t *testing.T) {
	cv := NewConfigValidator()

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	tests := []struct {
		name        string
		serviceName string
		expectErr   bool
	}{
		{
			name:        "valid service name",
			serviceName: "my-service",
			expectErr:   false,
		},
		{
			name:        "empty service name",
			serviceName: "",
			expectErr:   true,
		},
		{
			name:        "too long service name",
			serviceName: longName,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateServiceName(tt.serviceName)

			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateServiceName() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestConfigValidator_ValidateAttributes(t *testing.T) {
	cv := NewConfigValidator()

	// ValidateAttributes currently just returns nil, so it should never error
	err := cv.ValidateAttributes(nil)
	if err != nil {
		t.Errorf("ValidateAttributes() should not return error for nil, got %v", err)
	}

	err = cv.ValidateAttributes([]interface{}{"key", "value"})
	if err != nil {
		t.Errorf("ValidateAttributes() should not return error for valid attributes, got %v", err)
	}
}
