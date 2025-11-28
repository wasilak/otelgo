package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestTLSConfigBuildTLSConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *TLSConfig
		wantErr   bool
		checkFunc func(*testing.T, *tls.Config)
	}{
		{
			name: "default secure config",
			config: &TLSConfig{
				Insecure: false,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tc *tls.Config) {
				if tc.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify=false")
				}
			},
		},
		{
			name: "insecure config",
			config: &TLSConfig{
				Insecure: true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tc *tls.Config) {
				if !tc.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify=true")
				}
			},
		},
		{
			name: "with server name",
			config: &TLSConfig{
				ServerName: "example.com",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tc *tls.Config) {
				if tc.ServerName != "example.com" {
					t.Errorf("expected ServerName=example.com, got %s", tc.ServerName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.BuildTLSConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

func TestTLSConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *TLSConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid empty config",
			config:  &TLSConfig{Insecure: false},
			wantErr: false,
		},
		{
			name:    "valid insecure config",
			config:  &TLSConfig{Insecure: true},
			wantErr: false,
		},
		{
			name:    "conflicting insecure and CA",
			config:  &TLSConfig{Insecure: true, CACertPath: "/path/to/ca.pem"},
			wantErr: true,
			errMsg:  "cannot specify both Insecure=true and CACertPath",
		},
		{
			name:    "missing client key",
			config:  &TLSConfig{ClientCertPath: "/path/to/cert.pem"},
			wantErr: true,
			errMsg:  "both ClientCertPath and ClientKeyPath must be specified",
		},
		{
			name:    "missing client cert",
			config:  &TLSConfig{ClientKeyPath: "/path/to/key.pem"},
			wantErr: true,
			errMsg:  "both ClientCertPath and ClientKeyPath must be specified",
		},
		{
			name: "valid client cert pair",
			config: &TLSConfig{
				ClientCertPath: "/path/to/cert.pem",
				ClientKeyPath:  "/path/to/key.pem",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestNewTLSConfig(t *testing.T) {
	t.Run("from environment", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
		t.Setenv("OTEL_EXPORTER_OTLP_CERTIFICATE", "/path/to/ca.pem")
		t.Setenv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE", "/path/to/cert.pem")
		t.Setenv("OTEL_EXPORTER_OTLP_CLIENT_KEY", "/path/to/key.pem")
		t.Setenv("OTEL_EXPORTER_OTLP_SERVER_NAME", "example.com")

		config := NewTLSConfig()

		if !config.Insecure {
			t.Error("expected Insecure=true")
		}
		if config.CACertPath != "/path/to/ca.pem" {
			t.Errorf("expected CACertPath=/path/to/ca.pem, got %s", config.CACertPath)
		}
		if config.ClientCertPath != "/path/to/cert.pem" {
			t.Errorf("expected ClientCertPath=/path/to/cert.pem, got %s", config.ClientCertPath)
		}
		if config.ClientKeyPath != "/path/to/key.pem" {
			t.Errorf("expected ClientKeyPath=/path/to/key.pem, got %s", config.ClientKeyPath)
		}
		if config.ServerName != "example.com" {
			t.Errorf("expected ServerName=example.com, got %s", config.ServerName)
		}
	})

	t.Run("defaults when env not set", func(t *testing.T) {
		t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "")
		t.Setenv("OTEL_EXPORTER_OTLP_CERTIFICATE", "")
		t.Setenv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE", "")
		t.Setenv("OTEL_EXPORTER_OTLP_CLIENT_KEY", "")
		t.Setenv("OTEL_EXPORTER_OTLP_SERVER_NAME", "")

		config := NewTLSConfig()

		if config.Insecure {
			t.Error("expected Insecure=false by default")
		}
		if config.CACertPath != "" {
			t.Error("expected empty CACertPath by default")
		}
		if config.ServerName != "" {
			t.Error("expected empty ServerName by default")
		}
	})
}

func TestTLSConfigBuildTLSConfigWithCA(t *testing.T) {
	tempDir := t.TempDir()

	caCert := createTestCert(t)
	caCertPath := tempDir + "/ca.pem"
	if err := os.WriteFile(caCertPath, caCert, 0600); err != nil {
		t.Fatalf("failed to write CA cert: %v", err)
	}

	config := &TLSConfig{
		CACertPath: caCertPath,
	}

	tlsConfig, err := config.BuildTLSConfig()
	if err != nil {
		t.Fatalf("BuildTLSConfig() error = %v", err)
	}

	if tlsConfig.RootCAs == nil {
		t.Error("expected RootCAs to be set")
	}
}

func TestTLSConfigBuildTLSConfigWithInvalidCAPath(t *testing.T) {
	config := &TLSConfig{
		CACertPath: "/nonexistent/path/ca.pem",
	}

	_, err := config.BuildTLSConfig()
	if err == nil {
		t.Error("expected error for nonexistent CA path")
	}
}

func TestTLSConfigBuildTLSConfigWithInvalidCAFormat(t *testing.T) {
	tempDir := t.TempDir()

	invalidCertPath := tempDir + "/invalid.pem"
	if err := os.WriteFile(invalidCertPath, []byte("invalid cert data"), 0600); err != nil {
		t.Fatalf("failed to write invalid cert: %v", err)
	}

	config := &TLSConfig{
		CACertPath: invalidCertPath,
	}

	_, err := config.BuildTLSConfig()
	if err == nil {
		t.Error("expected error for invalid CA format")
	}
}

func TestTLSConfigBuildTLSConfigWithClientCerts(t *testing.T) {
	tempDir := t.TempDir()

	certPem, keyPem := createTestKeyPair(t)

	certPath := tempDir + "/cert.pem"
	keyPath := tempDir + "/key.pem"

	if err := os.WriteFile(certPath, certPem, 0600); err != nil {
		t.Fatalf("failed to write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPem, 0600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	config := &TLSConfig{
		ClientCertPath: certPath,
		ClientKeyPath:  keyPath,
	}

	tlsConfig, err := config.BuildTLSConfig()
	if err != nil {
		t.Fatalf("BuildTLSConfig() error = %v", err)
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("expected 1 certificate, got %d", len(tlsConfig.Certificates))
	}
}

func TestTLSConfigBuildTLSConfigWithInvalidClientCerts(t *testing.T) {
	tempDir := t.TempDir()

	certPath := tempDir + "/cert.pem"
	keyPath := tempDir + "/key.pem"

	if err := os.WriteFile(certPath, []byte("invalid"), 0600); err != nil {
		t.Fatalf("failed to write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("invalid"), 0600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	config := &TLSConfig{
		ClientCertPath: certPath,
		ClientKeyPath:  keyPath,
	}

	_, err := config.BuildTLSConfig()
	if err == nil {
		t.Error("expected error for invalid client certs")
	}
}

func createTestCert(t *testing.T) []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-ca",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
}

func createTestKeyPair(t *testing.T) ([]byte, []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-cert",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPem, keyPem
}
