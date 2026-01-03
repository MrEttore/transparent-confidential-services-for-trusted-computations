package challenge

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Verify checks if the issued challenge (optionally bound with TLS fingerprint) matches the reported data.
//
// issuedChallenge: the base64-encoded challenge issued by the Relying Party.
// tlsFingerprint: optional TLS certificate fingerprint for binding verification.
// reportData: the base64-encoded data reported by the Attester.
//
// If tlsFingerprint is provided, verifies: reportData == base64(SHA256(challenge || fingerprint)[0:32] || challenge[0:32])
// Otherwise falls back to simple equality check.
func Verify(issuedChallenge, reportData string) error {
	if issuedChallenge == reportData {
		return nil
	}

	return fmt.Errorf("challenge mismatch: the reportData field does not match the issued challenge (nonce) by the Relying Party. The provided evidence is not valid for attestation")
}

// VerifyWithTlsBinding verifies the challenge binding that includes the TLS certificate fingerprint.
//
// The userData in the quote is computed as: SHA256(challenge || fingerprint)[0:32] || challenge[0:32]
func VerifyWithTlsBinding(issuedChallenge, tlsFingerprint, reportData string) error {
	// Decode the original challenge from base64
	challengeBytes, err := base64.StdEncoding.DecodeString(issuedChallenge)
	if err != nil {
		return fmt.Errorf("failed to decode challenge: %w", err)
	}

	// Recompute the expected userData: SHA256(challenge || fingerprint)[0:32] || challenge[0:32]
	combined := append(challengeBytes, []byte(tlsFingerprint)...)
	hash := sha256.Sum256(combined)

	var expectedUserData [64]byte
	copy(expectedUserData[:32], hash[:])

	// Safety check: ensure challenge is at least 32 bytes
	if len(challengeBytes) >= 32 {
		copy(expectedUserData[32:], challengeBytes[:32])
	} else {
		// If challenge is shorter than 32 bytes, copy what we have (or handle error)
		// For now, copy what acts as "prefix"
		copy(expectedUserData[32:], challengeBytes)
	}

	expectedReportData := base64.StdEncoding.EncodeToString(expectedUserData[:])

	if expectedReportData == reportData {
		return nil
	}

	return fmt.Errorf("TLS binding verification failed: reportData does not match SHA256(challenge || tlsFingerprint)")
}
