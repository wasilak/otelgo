# Security Guide for otelgo

This document provides comprehensive security guidelines for using the otelgo library, including TLS configuration, certificate management, and best practices for secure OpenTelemetry operations.

## Table of Contents

1. [TLS Configuration](#tls-configuration)
2. [Certificate Setup Procedures](#certificate-setup-procedures)
3. [Security Best Practices](#security-best-practices)
4. [Threat Model & Mitigations](#threat-model--mitigations)
5. [Environment Variables](#environment-variables)

## TLS Configuration

The otelgo library provides secure TLS configuration across all OpenTelemetry components (logs, metrics, tracing) with proper validation and error handling.

### Basic TLS Configuration

All packages (logs, metrics, tracing) support TLS configuration through the `TLS` field in their respective configuration structs:

```go
import (
    "context"
    "github.com/wasilak/otelgo/logs"
    "github.com/wasilak/otelgo/internal"
)

config := logs.OtelGoLogsConfig{
    TLS: &internal.TLSConfig{
        Insecure:   false,                    // Set to true only for development
        CACertPath: "/path/to/ca-cert.pem",  // Path to CA certificate
        ServerName: "my-otel-collector.com", // Override server name for SNI
    },
}

ctx, provider, err := logs.Init(ctx, config)
if err != nil {
    log.Fatal(err)
}
```

### Advanced TLS Configuration

For mutual TLS (client certificate authentication):

```go
config := logs.OtelGoLogsConfig{
    TLS: &internal.TLSConfig{
        ClientCertPath: "/path/to/client-cert.pem",
        ClientKeyPath:  "/path/to/client-key.pem",
        CACertPath:     "/path/to/ca-cert.pem",
        ServerName:     "my-otel-collector.com",
    },
}
```

## Certificate Setup Procedures

### 1. Self-Signed Certificate Generation (Development Only)

For development purposes, you can generate self-signed certificates:

```bash
# Generate private key
openssl genrsa -out client-key.pem 2048

# Generate certificate signing request
openssl req -new -key client-key.pem -out client.csr

# Generate self-signed certificate
openssl x509 -req -in client.csr -signkey client-key.pem -out client-cert.pem -days 365
```

### 2. Production Certificate Setup

For production environments:

1. **Obtain certificates from a trusted CA** or use an internal PKI infrastructure
2. **Store certificates securely** with appropriate file permissions (600 recommended)
3. **Never commit certificates to source control**
4. **Set up certificate rotation procedures**

### 3. Certificate Validation

The library performs automatic validation:

- **Certificate chain verification** against provided CA certificates
- **Hostname verification** using provided ServerName
- **Expiration date checks** (handled by Go's crypto/tls package)
- **Proper certificate format validation**

## Security Best Practices

### 1. Environment Variables

Use environment variables for configuration:

| Environment Variable | Description | Example |
|---------------------|-------------|---------|
| `OTEL_EXPORTER_OTLP_INSECURE` | Set to "true" to disable TLS (development only) | `false` |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` | Path to CA certificate file | `/etc/ssl/certs/ca.crt` |
| `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` | Path to client certificate file | `/etc/ssl/certs/client.crt` |
| `OTEL_EXPORTER_OTLP_CLIENT_KEY` | Path to client key file | `/etc/ssl/private/client.key` |
| `OTEL_EXPORTER_OTLP_SERVER_NAME` | Override server name for TLS verification | `collector.example.com` |

### 2. Production Security Checklist

- [ ] **Never use `OTEL_EXPORTER_OTLP_INSECURE=true` in production**
- [ ] **Use certificates from trusted Certificate Authorities**
- [ ] **Set secure file permissions (600) on certificate and key files**
- [ ] **Enable mutual TLS authentication in production**
- [ ] **Verify server certificate against known CA**
- [ ] **Use specific server names for SNI validation**
- [ ] **Implement certificate rotation procedures**
- [ ] **Monitor for certificate expiration**

### 3. Network Security

- **Isolate collector traffic** on dedicated networks when possible
- **Use VPN or VPC peering** for collector communication
- **Implement firewall rules** limiting access to collectors
- **Monitor for suspicious traffic patterns**

### 4. Error Handling

Always properly handle TLS errors:

```go
// Good: Proper error handling
ctx, provider, err := logs.Init(ctx, config)
if err != nil {
    log.Printf("Failed to initialize logs: %v", err)
    // Handle error appropriately
    return err
}

// Avoid: Ignoring TLS errors
ctx, provider, _ := logs.Init(ctx, config) // Don't do this!
```

## Threat Model & Mitigations

### Threat: Man-in-the-Middle (MITM) Attacks

**Risk:** An attacker intercepting communication between clients and OpenTelemetry collectors.

**Mitigation:**
- Always use TLS with proper certificate verification
- Never use insecure connections in production
- Validate server certificates against known CAs
- Use certificate pinning if applicable

### Threat: Data Interception

**Risk:** Sensitive telemetry data being intercepted.

**Mitigation:**
- Encrypt all communication with TLS
- Use strong cipher suites (handled by Go's crypto/tls)
- Implement data minimization (only send necessary telemetry)

### Threat: Certificate Compromise

**Risk:** Stolen client certificates allowing unauthorized data submission.

**Mitigation:**
- Use short-lived certificates if possible
- Implement strict access controls to certificate files
- Enable mutual TLS only where necessary
- Regular certificate rotation

### Threat: Information Disclosure

**Risk:** Error messages or logs revealing sensitive configuration.

**Mitigation:**
- Avoid logging sensitive configuration values
- Use generic error messages in production
- Sanitize error logs for sensitive information

## Security Configuration Examples

### Secure Production Configuration

```go
config := logs.OtelGoLogsConfig{
    TLS: &internal.TLSConfig{
        Insecure:       false,
        CACertPath:     "/secure-path/ca-cert.pem",
        ClientCertPath: "/secure-path/client-cert.pem",
        ClientKeyPath:  "/secure-path/client-key.pem",
        ServerName:     "production-collector.example.com",
    },
    Attributes: []attribute.KeyValue{
        attribute.String("service.version", "1.0.0"),
        attribute.String("environment", "production"),
    },
}
```

### Development Configuration

```go
config := logs.OtelGoLogsConfig{
    TLS: &internal.TLSConfig{
        Insecure: true, // Only in development!
    },
}
```

## Common Security Issues and Solutions

### Issue: Certificate Verification Errors

**Problem:** `x509: certificate is not valid for any names`

**Solution:** Use the `ServerName` field to specify the correct hostname for SNI.

### Issue: Connection Refused During Certificate Validation

**Problem:** TLS handshake fails during validation

**Solution:** Verify that the CA certificate path is correct and the certificate is not expired.

### Issue: File Permission Vulnerabilities

**Problem:** Certificate files accessible to unauthorized users

**Solution:** Set file permissions to 600 and ensure proper ownership.

## Security Monitoring

Consider implementing monitoring for:

- **TLS handshake failures**
- **Unexpected connection patterns**
- **Certificate renewal alerts**
- **Unauthorized access attempts**

## Reporting Security Issues

If you discover a security vulnerability, please report it responsibly by contacting the maintainers directly rather than opening a public issue.

---

*This guide was last updated for otelgo version 1.0.0. Always refer to the latest documentation for the most current security recommendations.*