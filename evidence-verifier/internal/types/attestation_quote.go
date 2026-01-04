package types

// ### Request Types ###

// VerifyTdxQuoteRequest represents the JSON request payload for verifying a TDX quote.
// It contains a Quote object holding the attestation quote details.
//
//	{
//	  "quote": { ... }
//	}
type VerifyTdxQuoteRequest struct {
	IssuedChallenge           string `json:"issuedChallenge" validate:"required"`
	BaselineManifestUrl       string `json:"baselineManifestUrl" validate:"required,url"`
	TlsCertificateFingerprint string `json:"tlsCertificateFingerprint"`
	Quote                     Quote  `json:"quote" validate:"required"`
}

// Quote represents the TDX attestation quote structure
// that is sent for verification.
type Quote struct {
	ExtraBytes     string      `json:"extraBytes"`
	Header         QuoteHeader `json:"header" validate:"required"`
	SignedData     SignedData  `json:"signedData" validate:"required"`
	SignedDataSize int         `json:"signedDataSize" validate:"required"`
	TdQuoteBody    TdQuoteBody `json:"tdQuoteBody" validate:"required"`
}

// QuoteHeader contains metadata about the quote
// such as SVN values, vendor ID, and version.
type QuoteHeader struct {
	AttestationKeyType int    `json:"attestationKeyType"`
	PceSvn             string `json:"pceSvn"`
	QeSvn              string `json:"qeSvn"`
	QeVendorId         string `json:"qeVendorId"`
	TeeType            int    `json:"teeType"`
	UserData           string `json:"userData"`
	Version            int    `json:"version"`
}

// SignedData encapsulates certification info, the ECDSA key, and signature.
type SignedData struct {
	CertificationData   CertificationData `json:"certificationData"`
	EcdsaAttestationKey string            `json:"ecdsaAttestationKey"`
	Signature           string            `json:"signature"`
}

// TdQuoteBody mirrors the body portion of the quote for quick access.
type TdQuoteBody struct {
	MrConfigId     string   `json:"mrConfigId"`
	MrOwner        string   `json:"mrOwner"`
	MrOwnerConfig  string   `json:"mrOwnerConfig"`
	MrSeam         string   `json:"mrSeam"`
	MrSignerSeam   string   `json:"mrSignerSeam"`
	MrTd           string   `json:"mrTd"`
	ReportData     string   `json:"reportData"`
	Rtmrs          []string `json:"rtmrs"`
	SeamAttributes string   `json:"seamAttributes"`
	TdAttributes   string   `json:"tdAttributes"`
	TeeTcbSvn      string   `json:"teeTcbSvn"`
	Xfam           string   `json:"xfam"`
}

// CertificationData holds the certificate chain and related report data.
type CertificationData struct {
	CertificateDataType       int              `json:"certificateDataType"`
	QeReportCertificationData QeReportCertData `json:"qeReportCertificationData"`
	Size                      int              `json:"size"`
}

// QeReportCertData groups PCK chain, QE auth data, report, and its signature.
type QeReportCertData struct {
	PckCertificateChainData PckCertificateChainData `json:"pckCertificateChainData"`
	QeAuthData              QeAuthData              `json:"qeAuthData"`
	QeReport                QeReport                `json:"qeReport"`
	QeReportSignature       string                  `json:"qeReportSignature"`
}

// PckCertificateChainData carries the certificate chain and its size.
type PckCertificateChainData struct {
	CertificateDataType int    `json:"certificateDataType"`
	PckCertChain        string `json:"pckCertChain"`
	Size                int    `json:"size"`
}

// QeAuthData represents the authentication data and its parsed size.
type QeAuthData struct {
	Data           string `json:"data"`
	ParsedDataSize int    `json:"parsedDataSize"`
}

// QeReport contains the enclave report details and reserved fields.
type QeReport struct {
	Attributes string `json:"attributes"`
	CpuSvn     string `json:"cpuSvn"`
	IsvProdId  int    `json:"isvProdId"`
	IsvSvn     int    `json:"isvSvn"`
	MrEnclave  string `json:"mrEnclave"`
	MrSigner   string `json:"mrSigner"`
	ReportData string `json:"reportData"`
	Reserved1  string `json:"reserved1"`
	Reserved2  string `json:"reserved2"`
	Reserved3  string `json:"reserved3"`
	Reserved4  string `json:"reserved4"`
}

// ### Response Types ###

// VerifyTdxQuoteResponse is the JSON envelope returned by the VerifyTdxQuote handler.
//
// Status will be "success" or "error".
// Data, when present, holds the VerificationResultTdxQuote containing verification status and log.
// Message is an optional human-readable error when Status=="error".
type VerifyTdxQuoteResponse struct {
	Status  string             `json:"status"`
	Data    VerificationReport `json:"data,omitempty"`
	Message string             `json:"message,omitempty"`
}
