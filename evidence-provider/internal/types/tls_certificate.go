package types

// ### Request Types ###

type GetTlsCertificateRequest struct {
	Challenge string `json:"challenge"`
}

// ### Response Types ###

type GetTlsCertificateResponse struct {
	Status  string                 `json:"status"`
	Data    TlsCertificateEvidence `json:"data"`
	Message string                 `json:"message,omitempty"`
}

type TlsCertificateEvidence struct {
	CertificateFingerprint string `json:"certificateFingerprint"` // SHA256 of the DER-encoded certificate
	CertificatePEM         string `json:"certificatePem"`         // PEM-encoded certificate
	ReportData             string `json:"reportData"`             // hash(challenge || fingerprint) for nonce binding
}
