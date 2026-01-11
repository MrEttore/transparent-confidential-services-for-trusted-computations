package tlscertificate

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"evidence-provider/internal/types"
)

var (
	certFingerprint string
	certPEM         []byte
	once            sync.Once
	certDir         = "/app/ssl"
)

// InitCertificate generates a fresh self-signed TLS certificate at startup.
// Must be called once before FetchEvidence.
func InitCertificate() error {
	var initErr error
	once.Do(func() {
		initErr = generateAndStore()
	})
	return initErr
}

// FetchEvidence returns TLS certificate evidence bound to the challenge.
func FetchEvidence(ctx context.Context, challenge string) (types.TlsCertificateEvidence, error) {
	if certFingerprint == "" {
		return types.TlsCertificateEvidence{}, fmt.Errorf("TLS certificate not initialized")
	}

	// Bind the challenge to the certificate fingerprint: hash(challenge || fingerprint)
	reportData := hashWithNonce(challenge)

	return types.TlsCertificateEvidence{
		CertificateFingerprint: certFingerprint,
		CertificatePEM:         string(certPEM),
		ReportData:             reportData,
	}, nil
}

// GetFingerprint returns the SHA256 fingerprint of the current certificate.
func GetFingerprint() string {
	return certFingerprint
}

// generateAndStore creates a self-signed TLS certificate and corresponding ECDSA P-256
// private key, stores them on disk, and caches the certificate PEM and SHA-256
// fingerprint in package-level variables.
//
// It ensures certDir exists, generates a new key pair, builds a certificate template
// for "localhost" (also valid for "evidence-provider"), signs the certificate with
// itself, then writes:
//   - cert.pem (0644) containing the certificate in PEM format
//   - key.pem  (0600) containing the private key in PEM format
//
// On success it logs the paths and a shortened fingerprint. Returns an error if
// any filesystem or crypto operation fails.
func generateAndStore() error {
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create ssl directory: %w", err)
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Evidence Provider"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "evidence-provider"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	fingerprint := sha256.Sum256(certDER)
	certFingerprint = hex.EncodeToString(fingerprint[:])

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	certPath := filepath.Join(certDir, "cert.pem")
	keyPath := filepath.Join(certDir, "key.pem")

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	log.Printf("TLS certificate generated: fingerprint=%s", certFingerprint[:16]+"...")
	log.Printf("Certificate: %s", certPath)
	log.Printf("Private key: %s", keyPath)

	return nil
}

// hashWithNonce returns a SHA-256 hex digest computed from the concatenation of the
// provided challenge (nonce) and the package-level certFingerprint.
// The output is a lowercase hexadecimal string suitable for use as a deterministic
// proof-of-possession value bound to both the nonce and the certificate fingerprint.
func hashWithNonce(challenge string) string {
	data := challenge + certFingerprint
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
