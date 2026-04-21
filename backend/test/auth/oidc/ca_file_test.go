package auth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/config"
	"github.com/flatcar/nebraska/backend/pkg/server"
	"github.com/flatcar/nebraska/backend/pkg/tlsutil"
)

type testCA struct {
	certPEM   []byte
	serverTLS *tls.Config
}

func newTestCA(t *testing.T, addr string) *testCA {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	host, _, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP(host)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	return &testCA{
		certPEM:   certPEM,
		serverTLS: &tls.Config{Certificates: []tls.Certificate{tlsCert}},
	}
}

func (ca *testCA) writeCAFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "ca-*.pem")
	require.NoError(t, err)
	_, err = f.Write(ca.certPEM)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func startTLSOIDCServer(t *testing.T) (*mockoidc.MockOIDC, *testCA) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	ca := newTestCA(t, ln.Addr().String())
	tlsLn := tls.NewListener(ln, ca.serverTLS)

	m, err := mockoidc.NewServer(nil)
	require.NoError(t, err)
	err = m.Start(tlsLn, ca.serverTLS)
	require.NoError(t, err)

	return m, ca
}

func TestCAFileOIDCSetup(t *testing.T) {
	t.Run("tls_oidc_with_ca_file_succeeds", func(t *testing.T) {
		oidcServer, ca := startTLSOIDCServer(t)
		defer oidcServer.Shutdown() //nolint:errcheck

		var testConfig config.Config
		err := copier.Copy(&testConfig, conf)
		require.NoError(t, err)

		caFile := ca.writeCAFile(t)
		testConfig.CAFile = caFile
		testConfig.OidcIssuerURL = oidcServer.Issuer()

		caPool, err := tlsutil.LoadCAPool(caFile)
		require.NoError(t, err)
		testConfig.CACertPool = caPool

		db := newDBForTest(t)
		srv, err := server.New(&testConfig, db)
		assert.NotNil(t, srv)
		assert.NoError(t, err)
	})

	t.Run("tls_oidc_without_ca_file_fails", func(t *testing.T) {
		oidcServer, _ := startTLSOIDCServer(t)
		defer oidcServer.Shutdown() //nolint:errcheck

		var testConfig config.Config
		err := copier.Copy(&testConfig, conf)
		require.NoError(t, err)

		testConfig.OidcIssuerURL = oidcServer.Issuer()

		db := newDBForTest(t)
		srv, err := server.New(&testConfig, db)
		assert.Nil(t, srv)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error setting up oidc provider")
		assert.Contains(t, err.Error(), "certificate signed by unknown authority")
	})
}
