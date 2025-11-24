package decubcrypto

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// LoadTLSCertificates loads client or server certificates from files
func LoadTLSCertificates(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load TLS certificates: %w", err)
	}
	return cert, nil
}

// LoadCACertificate loads a CA certificate from file
func LoadCACertificate(caFile string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return caCertPool, nil
}

// CreateClientTLSConfig creates a TLS config for client connections with mTLS
func CreateClientTLSConfig(clientCert tls.Certificate, caCertPool *x509.CertPool) *tls.Config {
	return &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:           caCertPool,
		InsecureSkipVerify: false, // Always verify server certificate
		MinVersion:        tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}
}

// CreateServerTLSConfig creates a TLS config for server with mTLS requirement
func CreateServerTLSConfig(serverCert tls.Certificate, caCertPool *x509.CertPool) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert, // Require client certificates
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}
}

// VerifyMutualTLS performs additional verification for mutual TLS
func VerifyMutualTLS(conn *tls.Conn) error {
	state := conn.ConnectionState()
	if !state.HandshakeComplete {
		return fmt.Errorf("TLS handshake not complete")
	}

	// Verify client certificate was provided and verified
	if len(state.PeerCertificates) == 0 {
		return fmt.Errorf("no client certificate provided")
	}

	// Additional custom verification can be added here
	// For example, checking certificate fields, revocation lists, etc.

	return nil
}
