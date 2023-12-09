package crypto

import (
	"bytes"
	secureRand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/pygrum/monarch/pkg/config"
)

type CertVerifier func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

var (
	r *rand.Rand
)

func init() {
	source := rand.NewSource(time.Now().Unix())
	r = rand.New(source)
}

func PeerCertificateVerifier(caCertPEM []byte) CertVerifier {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		certs := make([]*x509.Certificate, len(rawCerts))
		for i, asn1Data := range rawCerts {
			cert, err := x509.ParseCertificate(asn1Data)
			if err != nil {
				return fmt.Errorf("failed to parse certificate: %v", err)
			}
			certs[i] = cert
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCertPEM)
		opts := x509.VerifyOptions{
			Roots:         certPool,
			CurrentTime:   time.Now(),
			DNSName:       "", // Skip hostname verification
			Intermediates: x509.NewCertPool(),
		}

		for i, cert := range certs {
			if i == 0 {
				continue
			}
			opts.Intermediates.AddCert(cert)
		}
		_, err := certs[0].Verify(opts)
		return err
	}
}

// NewClientCertificate generates a cert-key pair for a newly created operator
func NewClientCertificate(cn string) ([]byte, []byte, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(int64(randInt(randInt(5000)))), // lol
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	caCert, caKey, err := CertificateAuthority()
	if err != nil {
		return nil, nil, err
	}
	privKey, err := rsa.GenerateKey(secureRand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	// create and sign with ca priv key (caKey)
	certData, err := x509.CreateCertificate(secureRand.Reader, cert, caCert, &privKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certData,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}

func CertificateAuthority() (*x509.Certificate, *rsa.PrivateKey, error) {
	encodedCert, encodedKey, err := CaCertKeyPair()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get CA cert data: %v", err)
	}
	certBlock, _ := pem.Decode(encodedCert)
	if certBlock == nil {
		return nil, nil, errors.New("no ca cert PEM data found")
	}
	keyBlock, _ := pem.Decode(encodedKey)
	if keyBlock == nil {
		return nil, nil, errors.New("no ca key PEM data found")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't parse CA certificate: %v", err)
	}
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't parse CA key: %v", err)
	}
	return cert, key.(*rsa.PrivateKey), nil
}

func ClientTLSConfig(c *config.MonarchClientConfig) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(c.CertPEM, c.KeyPEM)
	if err != nil {
		return nil, fmt.Errorf("couldn't create certificate key pair: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(c.CaCertPEM)
	tlsConfig := &tls.Config{
		Certificates:          []tls.Certificate{cert},
		RootCAs:               caCertPool,
		InsecureSkipVerify:    true,
		VerifyPeerCertificate: PeerCertificateVerifier(c.CaCertPEM),
	}
	return tlsConfig, nil
}

// CaCertKeyPair returns pem encoded certificate and key for monarchCA
func CaCertKeyPair() ([]byte, []byte, error) {
	cert, err := os.ReadFile(config.MainConfig.CaCert)
	if err != nil {
		return nil, nil, err
	}
	key, err := os.ReadFile(config.MainConfig.CaKey)
	if err != nil {
		return nil, nil, err
	}
	return cert, key, nil
}

func randInt(upper int) int {
	return r.Intn(upper)
}
