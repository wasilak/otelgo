# Troubleshooting Guide for otelgo

This document provides common issues and solutions for troubleshooting the otelgo library. It covers debugging techniques, frequent error messages, and solutions for common setup problems.

## Table of Contents

1. [Getting Started with Debugging](#getting-started-with-debugging)
2. [Common Issues and Solutions](#common-issues-and-solutions)
3. [Error Messages and Solutions](#error-messages-and-solutions)
4. [Debugging Techniques](#debugging-techniques)
5. [Performance Issues](#performance-issues)
6. [Network Connectivity Problems](#network-connectivity-problems)
7. [TLS/SSL Configuration Issues](#tlsssl-configuration-issues)

## Getting Started with Debugging

Before diving into specific troubleshooting, gather information about your issue:

- **otelgo version**: Check your current version
- **Go version**: Some issues might be version-specific
- **Environment**: Development, staging, or production
- **Network setup**: Local, Docker, Kubernetes, cloud provider
- **Collector type**: OpenTelemetry Collector, Jaeger, Zipkin, etc.
- **Error messages**: Exact error text and context

## Common Issues and Solutions

### 1. No Data Being Exported

**Symptoms:**
- Telemetry data is generated but not appearing in the collector
- No errors are thrown despite using telemetry

**Solutions:**
1. **Check endpoint configuration**: Ensure the OTLP endpoint URL is correct
   ```bash
   export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="https://my-collector:4318"
   export OTEL_EXPORTER_OTLP_LOGS_ENDPOINT="https://my-collector:4318/v1/logs"
   export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT="https://my-collector:4318/v1/metrics"
   ```
   
2. **Verify protocol matching**: Ensure protocol matches the collector's listening port
   ```bash
   export OTEL_EXPORTER_OTLP_TRACES_PROTOCOL="http"  # Use 'grpc' if collector uses gRPC
   export OTEL_EXPORTER_OTLP_LOGS_PROTOCOL="http"
   export OTEL_EXPORTER_OTLP_METRICS_PROTOCOL="http"
   ```

3. **Check firewall/networking**: Ensure the collector port is accessible
   ```bash
   # Test connectivity
   curl -v https://my-collector:4318/v1/traces
   ```

4. **Look for silent failures**: Ensure shutdown methods are called
   ```go
   defer logs.Shutdown(ctx, provider)  // Important for flushing data
   defer metrics.Shutdown(ctx, provider)
   defer tracing.Shutdown(ctx, provider)
   ```

### 2. TLS/SSL Connection Failures

**Symptoms:**
- `x509: certificate signed by unknown authority`
- `tls: failed to verify certificate`
- `connection reset by peer`

**Solutions:**
1. **For development only**: Use insecure connection (NEVER in production)
   ```go
   // Only for development/debugging!
   config := logs.OtelGoLogsConfig{
       TLS: &internal.TLSConfig{Insecure: true},
   }
   ```

2. **Provide proper CA certificate**:
   ```go
   config := logs.OtelGoLogsConfig{
       TLS: &internal.TLSConfig{
           CACertPath: "/path/to/ca-cert.pem",  // Path to CA cert
           ServerName: "collector.example.com",  // Required if using IP address
       },
   }
   ```

3. **Verify certificate validity**:
   ```bash
   # Check certificate expiration
   openssl x509 -in /path/to/cert.pem -text -noout
   
   # Verify against CA
   openssl verify -CAfile /path/to/ca-cert.pem /path/to/cert.pem
   ```

### 3. Environment Variable Issues

**Symptoms:**
- Configuration from code is ignored
- Unexpected default behavior
- Different behavior between environments

**Solutions:**
1. **Check environment variable precedence**:
   - Environment variables take precedence over programmatic configuration
   - Clear unwanted environment variables for testing: `unset OTEL_EXPORTER_OTLP_INSECURE`

2. **Verify variable names**:
   ```bash
   # Check current values
   env | grep OTEL
   
   # Set required variables
   export OTEL_SERVICE_NAME="my-service"
   export OTEL_EXPORTER_OTLP_INSECURE="false"
   ```

### 4. Builder Pattern Issues

**Symptoms:**
- Builder methods don't seem to affect behavior
- Unexpected configuration values
- Compile errors with builder methods

**Solutions:**
1. **Ensure proper chaining**:
   ```go
   // Correct usage - each method returns the builder
   ctx, provider, err := logs.NewBuilder().
       WithAttributes(attribute.String("env", "production")).
       WithTLS(&internal.TLSConfig{
           Insecure: true,  // Development only!
       }).
       Build(ctx)
   ```

2. **Verify import paths**:
   ```go
   import "github.com/wasilak/otelgo/logs"  // Make sure import is correct
   import "github.com/wasilak/otelgo/internal"
   ```

## Error Messages and Solutions

### Error: `Post "https://collector:4318/v1/traces": dial tcp: lookup collector: no such host`

**Cause:** DNS resolution failure - collector hostname cannot be resolved.

**Solutions:**
- Use IP address instead of hostname: `https://127.0.0.1:4318`
- Verify DNS configuration
- Check if collector is running and accessible

### Error: `failed to upload.*: Post "...": dial tcp: connection refused`

**Cause:** Collector is not listening on the specified port.

**Solutions:**
1. Verify collector is running:
   ```bash
   # Check if process is running
   ps aux | grep otel-collector
   ```

2. Check correct port:
   ```bash
   # Check listening ports
   netstat -an | grep 4318
   lsof -i :4318  # macOS/Linux
   ```

3. Verify protocol matches (GRPC vs HTTP):
   - GRPC typically uses port 4317
   - HTTP typically uses port 4318

### Error: `x509: certificate relies on legacy Common Name`

**Cause:** Using outdated certificate without proper Subject Alternative Names (SAN).

**Solutions:**
1. **Regenerate certificate with SAN**:
   ```bash
   # Create certificate config file
   cat > cert.conf << EOF
   [req]
   distinguished_name = req_distinguished_name
   x509_extensions = v3_req
   prompt = no
   
   [req_distinguished_name]
   CN=localhost
   
   [v3_req]
   subjectAltName = @alt_names
   
   [alt_names]
   DNS.1 = localhost
   IP.1 = 127.0.0.1
   EOF
   
   # Generate certificate with SAN
   openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -config cert.conf
   ```

### Error: `context deadline exceeded`

**Cause:** Timeout during connection establishment or data transmission.

**Solutions:**
1. **Increase connection timeout** (implementation dependent)
2. **Check network connectivity**
3. **Reduce data payload size temporarily**
4. **Verify collector performance under load**

### Error: `invalid TLS configuration`

**Cause:** Conflicting TLS configuration (e.g., both `Insecure: true` and CA certificate path).

**Solutions:**
1. **Review TLS configuration**:
   ```go
   // Don't do this:
   &internal.TLSConfig{
       Insecure: true,
       CACertPath: "/some/path",  // This conflicts with Insecure: true
   }
   
   // Do this instead:
   &internal.TLSConfig{
       Insecure: true,
       // Don't set CACertPath when Insecure is true
   }
   // OR
   &internal.TLSConfig{
       CACertPath: "/proper/ca/path",  // Only if Insecure is false
       Insecure: false,
   }
   ```

## Debugging Techniques

### 1. Enable Verbose Logging

To debug connection and configuration issues:

```go
import (
    "log"
    "context"
    "github.com/wasilak/otelgo/logs"
)

// Use a verbose logger to see what's happening
logger := log.Default()
logger.SetOutput(os.Stdout)

// Initialize with common configuration
config := logs.OtelGoLogsConfig{
    // Your config
}

ctx := context.Background()
ctx, provider, err := logs.Init(ctx, config)
if err != nil {
    log.Printf("Error initializing logs: %v", err)
} else {
    log.Println("Successfully initialized logs provider")
    defer logs.Shutdown(ctx, provider)
}
```

### 2. Test Individual Components

Test each package separately to isolate issues:

```go
// Test logs only
ctx, logProvider, err := logs.NewBuilder().WithTLSInsecure().Build(ctx)
if err != nil {
    log.Printf("Logs initialization failed: %v", err)
} else {
    defer logs.Shutdown(ctx, logProvider)
}

// Test metrics only  
ctx, metricProvider, err := metrics.NewBuilder().WithTLSInsecure().Build(ctx)
if err != nil {
    log.Printf("Metrics initialization failed: %v", err)
} else {
    defer metrics.Shutdown(ctx, metricProvider)
}
```

### 3. Use Network Debugging Tools

For network connectivity issues:

```bash
# TCP connectivity
telnet collector-host 4318

# More detailed connectivity check
curl -v --http1.1 -X POST http://collector:4318/v1/traces --data '{}'

# Check SSL certificate details (if using HTTPS)
openssl s_client -connect collector:4318 -servername collector
```

### 4. Verify Configuration Values

Print configuration at runtime (be careful with sensitive data):

```go
// DEBUG ONLY - Never in production
fmt.Printf("Service name: %s\n", os.Getenv("OTEL_SERVICE_NAME"))
fmt.Printf("Logs endpoint: %s\n", os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"))
fmt.Printf("Protocol: %s\n", os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"))
// etc.
```

## Performance Issues

### High Memory Usage

**Symptoms:** Rapid increase in memory consumption.

**Solutions:**
1. **Call shutdown methods properly** to flush and cleanup resources:
   ```go
   defer logs.Shutdown(ctx, provider)
   defer metrics.Shutdown(ctx, provider)
   defer tracing.Shutdown(ctx, provider)
   ```

2. **Adjust batch sizes** if collector is slow:
   - Environment variables for batching behavior are specific to the SDK implementation

3. **Reduce attribute cardinality** (avoid high-cardinality attributes like user IDs)

### Slow Application Startup

**Symptoms:** Extended startup time when initializing otelgo components.

**Solutions:**
1. **Use background initialization** if possible
2. **Delay initialization** until after app is serving requests
3. **Use non-blocking initialization** where possible

## Network Connectivity Problems

### Connection Timeouts

**Solutions:**
1. **Verify collector health**:
   ```bash
   # Check if collector is responsive
   curl -s http://collector:8888/metrics  # If health check endpoint is enabled
   ```

2. **Check for intermittent connectivity**:
   ```bash
   # Continuous ping test
   ping collector-host
   ```

### Firewall Issues

**Symptoms:** Connection works locally but fails in deployed environment.

**Solutions:**
1. **Check firewall rules** for the collector ports (usually 4317 for gRPC, 4318 for HTTP)
2. **Verify security groups** in cloud environments
3. **Check proxy settings** if applicable

## TLS/SSL Configuration Issues

### Certificate Verification Errors

**Solutions:**
1. **Verify certificate chain**:
   ```bash
   # Check certificate details
   openssl x509 -in /path/to/certificate.pem -text -noout | grep -A 1 -B 1 "Subject Alternative Name"
   ```

2. **Check certificate validity period**:
   ```bash
   # Check if certificate is expired
   date -d @$(openssl x509 -in /path/to/cert.pem -text -noout | grep -A 1 "Not After" | tail -1 | awk '{print $4" "$5" "$6" "$7}')
   ```

### Mutual TLS Issues

**Symptoms:** `tls: client didn't provide a certificate`

**Solutions:**
1. **Provide both client certificate and key**:
   ```go
   config := logs.OtelGoLogsConfig{
       TLS: &internal.TLSConfig{
           ClientCertPath: "/path/to/client.cert",
           ClientKeyPath:  "/path/to/client.key",
           CACertPath:     "/path/to/ca.cert",
       },
   }
   ```

2. **Ensure proper file permissions**:
   ```bash
   chmod 600 /path/to/client.key  # Private key shouldn't be world-readable
   ```

---

## Support Resources

- **Issues**: Report bugs or issues on the GitHub repository
- **Documentation**: Full API documentation at pkg.go.dev
- **Examples**: Check the examples directory in the repository

For security-related issues, please contact the maintainers directly rather than opening a public issue.

---

*This troubleshooting guide was last updated for otelgo version 1.0.0. Always refer to the latest documentation for the most current troubleshooting tips and best practices.*