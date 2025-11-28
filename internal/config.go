package internal

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// ValidationError represents an error that occurs during configuration validation.
type ValidationError struct {
	Field string
	Value interface{}
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%v': %v", e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// ConfigValidator provides methods for validating various types of configuration.
type ConfigValidator struct{}

// ValidateEndpoint validates an endpoint URL.
func (cv *ConfigValidator) ValidateEndpoint(endpoint string) error {
	if endpoint == "" {
		return &ValidationError{
			Field: "endpoint",
			Value: endpoint,
			Err:   errors.New("endpoint is required"),
		}
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return &ValidationError{
			Field: "endpoint",
			Value: endpoint,
			Err:   fmt.Errorf("invalid URL format: %w", err),
		}
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" &&
		parsedURL.Scheme != "grpc" && parsedURL.Scheme != "grpcs" {
		return &ValidationError{
			Field: "endpoint",
			Value: endpoint,
			Err:   errors.New("endpoint must use http, https, grpc, or grpcs scheme"),
		}
	}

	return nil
}

// ValidateInterval validates a time interval for metrics collection.
func (cv *ConfigValidator) ValidateInterval(interval time.Duration) error {
	if interval <= 0 {
		return &ValidationError{
			Field: "interval",
			Value: interval,
			Err:   errors.New("interval must be positive"),
		}
	}

	// Add reasonable upper bound to prevent excessive delays
	if interval > 24*time.Hour {
		return &ValidationError{
			Field: "interval",
			Value: interval,
			Err:   errors.New("interval must not exceed 24 hours"),
		}
	}

	return nil
}

// ValidateTLSConfig validates TLS configuration.
func (cv *ConfigValidator) ValidateTLSConfig(tlsConfig *TLSConfig) error {
	if tlsConfig == nil {
		return &ValidationError{
			Field: "tls",
			Value: tlsConfig,
			Err:   errors.New("TLS configuration cannot be nil"),
		}
	}

	// Use the existing TLS validation
	return tlsConfig.Validate()
}

// ValidateAttributes validates attribute lists for common issues.
func (cv *ConfigValidator) ValidateAttributes(attrs []interface{}) error {
	if attrs == nil {
		return nil // nil attributes are valid
	}

	// In a real application, you might want to validate attribute format
	// For now, we just ensure it's not problematic

	// Could add checks here for attributes, but for now just return nil
	// since the OpenTelemetry SDK handles attribute validation internally
	return nil
}

// ValidateServiceName validates the service name.
func (cv *ConfigValidator) ValidateServiceName(serviceName string) error {
	if serviceName == "" {
		return &ValidationError{
			Field: "service_name",
			Value: serviceName,
			Err:   errors.New("service name is required"),
		}
	}

	// Service names should not be excessively long
	if len(serviceName) > 100 {
		return &ValidationError{
			Field: "service_name",
			Value: serviceName,
			Err:   errors.New("service name exceeds maximum length of 100 characters"),
		}
	}

	return nil
}

// NewConfigValidator creates a new instance of ConfigValidator.
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}
