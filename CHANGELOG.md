# Changelog

All notable changes to the otelgo library will be documented in this file.

## [v1.0.0] - 2025-11-28

### 🎉 Features

#### Security Enhancements
- **TLS Configuration**: Added comprehensive TLS configuration support across all packages (logs, metrics, tracing)
  - Secure-by-default TLS implementation with proper certificate validation
  - Support for custom CA certificates, client certificates (mTLS), and server name overrides
  - Removal of hardcoded insecure settings throughout the codebase
  - Validation to prevent conflicting TLS configurations (e.g., both Insecure=true and CA cert path)
- **Environment Variable Support**: Added support for standard OpenTelemetry environment variables
  - `OTEL_EXPORTER_OTLP_INSECURE` - Control TLS verification
  - `OTEL_EXPORTER_OTLP_CERTIFICATE` - Custom CA certificate path
  - `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` - Client certificate for mutual TLS
  - `OTEL_EXPORTER_OTLP_CLIENT_KEY` - Client private key for mutual TLS
  - `OTEL_EXPORTER_OTLP_SERVER_NAME` - Server name override for certificate verification

#### Builder Pattern API
- **Fluent Configuration**: Introduced builder patterns for all packages
  - `logs.NewBuilder()` - Fluent API for log configuration
  - `metrics.NewBuilder()` - Fluent API for metrics configuration
  - `tracing.NewBuilder()` - Fluent API for tracing configuration
  - Method chaining for clean, readable configuration code
- **Consistent Interface**: Standardized configuration API across all packages

#### Configuration & Validation
- **Centralized Validation**: Added comprehensive configuration validation module
  - Endpoint validation with proper URL format checking
  - TLS configuration validation with conflict detection
  - Interval validation for time-based settings
  - Service name validation with length limits
- **Backward Compatibility**: Maintained compatibility with existing configuration patterns

#### Host & Runtime Metrics
- **Enhanced Metrics**: Improved host and runtime metrics support in tracing package
  - Configurable intervals for both host and runtime metrics
  - Proper TLS configuration support for metrics exporters
  - Error handling improvements to prevent crashes

### 🔧 Improvements

#### Performance & Concurrency
- **Thread Safety**: Fixed global race conditions in initialization
  - All Init functions now use local configuration copies
  - Eliminated global state mutations during initialization
  - Safe for concurrent initialization from multiple goroutines
- **Memory Efficiency**: Optimized memory allocations during initialization
- **Error Handling**: Improved error handling throughout all packages
  - Proper error propagation instead of panics where appropriate
  - Clear error messages for debugging

#### Testing & Quality
- **Comprehensive Tests**: Added extensive test coverage
  - Unit tests for all packages and configurations
  - Race condition tests with concurrent initialization
  - Error path tests covering invalid configurations
  - TLS validation tests for all security scenarios
- **Benchmarking**: Added performance benchmarks for all packages
- **Code Quality**: Applied consistent formatting and linting

### 🛡️ Security Fixes

- **Hardcoded InsecureSkipVerify**: Removed all hardcoded `InsecureSkipVerify: true` settings
  - Previously found in tracing package and metrics host/runtimemetrics
  - Now properly configurable via TLS configuration
  - Secure-by-default behavior restored
- **Certificate Validation**: Implemented proper certificate validation
  - CA certificate support with proper verification
  - Server name verification for SNI support
  - Client certificate authentication (mTLS) support
- **Configuration Validation**: Added validation to prevent common security misconfigurations

### 🚀 Breaking Changes

- **TLS Configuration**: Default behavior is now secure (no more hardcoded InsecureSkipVerify)
  - Applications that relied on insecure connections must now explicitly configure `Insecure: true`
  - This is a security improvement, but may require configuration updates in development environments
- **API Changes**: Some direct access to global configuration variables might not work as expected due to local copy usage in Init functions

### 📚 Documentation

- **Security Guide**: Comprehensive security best practices and TLS configuration guide
- **Configuration Guide**: Detailed documentation for all configuration options
- **Troubleshooting Guide**: Common issues and solutions
- **API Reference**: Complete API documentation for all public interfaces
- **Examples**: Multiple working examples covering basic usage, TLS configurations, error handling, and concurrent initialization

### 🧪 Examples

- **Basic Example**: Simple initialization of all OpenTelemetry components
- **TLS Custom CA Example**: Secure TLS configuration with custom CA certificates
- **Client Certificate Example**: Mutual TLS setup with client certificates
- **Error Handling Example**: Proper error handling and fallback strategies
- **Concurrent Initialization Example**: Safe concurrent initialization patterns

### 🏗️ Internal Improvements

- **Modular Architecture**: Better separation of concerns between packages
- **Consistent Patterns**: Unified configuration and initialization patterns across all packages
- **Validation Module**: Centralized configuration validation for consistency
- **Testing Framework**: Comprehensive test suite covering all functionality

## v0.x.x - Previous Versions

Previous versions were internal development releases that focused on basic OpenTelemetry functionality. This v1.0.0 release represents the first production-ready version with comprehensive security, configuration, and documentation.

[Unreleased]: https://github.com/wasilak/otelgo/compare/v1.0.0...HEAD
[v1.0.0]: https://github.com/wasilak/otelgo/releases/tag/v1.0.0