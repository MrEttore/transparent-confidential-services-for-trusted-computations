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
	CertificateFingerprint string `json:"certificateFingerprint"`
	CertificatePEM         string `json:"certificatePem"`
	ReportData             string `json:"reportData"`
}
