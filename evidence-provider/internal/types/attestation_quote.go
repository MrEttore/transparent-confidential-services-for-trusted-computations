package types

// ### Request Types ###

type GetTdxQuoteRequest struct {
	Challenge string `json:"challenge"`
}

// ### Response Types ###

type GetTdxQuoteResponse struct {
	Status  string                   `json:"status"`
	Data    AttestationQuoteEvidence `json:"data"`
	Message string                   `json:"message,omitempty"`
}

type AttestationQuoteEvidence struct {
	Quote                     map[string]any `json:"quote"`
	TlsCertificateFingerprint string         `json:"tlsCertificateFingerprint,omitempty"` // SHA256 of TLS cert bound to quote
	UserDataComposition       string         `json:"userDataComposition,omitempty"`       // Explains how userData was computed
}
