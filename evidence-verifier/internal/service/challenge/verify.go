package challenge

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Verify checks if the issued challenge matches the quote's reportData field.
//
// - issuedChallenge: the base64-encoded challenge issued by the Relying Application.
//
// - reportData: the base64-encoded data reported by the Attester.
func Verify(issuedChallenge, reportData string) error {
	if issuedChallenge == reportData {
		return nil
	}

	return fmt.Errorf("challenge mismatch: the reportData field does not match the issued challenge (nonce) by the Relying Party. The provided evidence is not valid for attestation")
}

// VerifyWithTlsBinding verifies the challenge binding that includes the TLS certificate fingerprint.
//
// - issuedChallenge: the base64-encoded challenge issued by the Relying Application.
//
// - tlsFingerprint: optional TLS certificate fingerprint for binding verification.
//
// - reportData: the base64-encoded data reported by the Attester.
//
// The reportData field in the quote is computed as: SHA256(challenge || fingerprint[0:32] || challenge[0:32]
func VerifyWithTlsBinding(issuedChallenge, tlsFingerprint, reportData string) error {
	challengeBytes, err := base64.StdEncoding.DecodeString(issuedChallenge)
	if err != nil {
		return fmt.Errorf("failed to decode challenge: %w", err)
	}
	
	// Recompute the expected reportData value.
	combined := append(challengeBytes, []byte(tlsFingerprint)...)
	hash := sha256.Sum256(combined)

	var expectedReportData [64]byte
	copy(expectedReportData[:32], hash[:])

	if len(challengeBytes) >= 32 {
		copy(expectedReportData[32:], challengeBytes[:32])
	} else {
		copy(expectedReportData[32:], challengeBytes)
	}

	expectedReportDataStr := base64.StdEncoding.EncodeToString(expectedReportData[:])

	if expectedReportDataStr == reportData {
		return nil
	}

	return fmt.Errorf("TLS binding verification failed: reportData does not match SHA256(challenge || tlsFingerprint)")
}
