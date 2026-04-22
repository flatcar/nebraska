// Package tlsutil provides helpers for building HTTP clients with custom CA trust.
package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

// LoadCAPool reads PEM-encoded certificates from caFile and returns a cert pool
// containing both the system CAs and the custom ones.
// Returns nil, nil when caFile is empty.
func LoadCAPool(caFile string) (*x509.CertPool, error) {
	if caFile == "" {
		return nil, nil
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("error loading system CA pool: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("error reading ca-file %q: %w", caFile, err)
	}

	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("ca-file %q contains no valid PEM certificates", caFile)
	}

	return pool, nil
}

// NewHTTPClient creates an *http.Client that trusts the provided cert pool.
// Returns nil when pool is nil.
func NewHTTPClient(pool *x509.CertPool) *http.Client {
	if pool == nil {
		return nil
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		RootCAs: pool,
	}

	return &http.Client{Transport: transport}
}
