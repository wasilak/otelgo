package internal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type TLSConfig struct {
	Insecure       bool
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string
	ServerName     string
}

func NewTLSConfig() *TLSConfig {
	return &TLSConfig{
		Insecure:       os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true",
		CACertPath:     os.Getenv("OTEL_EXPORTER_OTLP_CERTIFICATE"),
		ClientCertPath: os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE"),
		ClientKeyPath:  os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_KEY"),
		ServerName:     os.Getenv("OTEL_EXPORTER_OTLP_SERVER_NAME"),
	}
}

// BuildTLSConfig creates a TLS configuration from the TLSConfig.
// IMPORTANT: InsecureSkipVerify is only set to true when c.Insecure is explicitly true.
// In production environments, c.Insecure should always be false and proper certificates should be configured.
func (c *TLSConfig) BuildTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.Insecure,
		ServerName:         c.ServerName,
	}

	if c.CACertPath != "" {
		caCert, err := os.ReadFile(c.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	if c.ClientCertPath != "" && c.ClientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(c.ClientCertPath, c.ClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func (c *TLSConfig) Validate() error {
	if c.Insecure && c.CACertPath != "" {
		return fmt.Errorf("cannot specify both Insecure=true and CACertPath")
	}

	if (c.ClientCertPath != "") != (c.ClientKeyPath != "") {
		return fmt.Errorf("both ClientCertPath and ClientKeyPath must be specified")
	}

	return nil
}
