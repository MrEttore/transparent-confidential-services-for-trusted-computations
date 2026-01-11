package handler

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"

	"evidence-provider/internal/service/attestationquote"
	"evidence-provider/internal/service/challenge"
	"evidence-provider/internal/service/tlscertificate"
	"evidence-provider/internal/types"
	"evidence-provider/internal/util"
)

// GetTdxQuote handles HTTP requests to fetch an Intel TDX attestation quote.
// It validates the incoming request, decodes the challenge, fetches the attestation evidence,
// and returns the quote data in the response.
//
// The reportData field in the quote contains: SHA256(challenge || tlsCertificateFingerprint)
// This binds both the client's nonce and the TLS certificate to the hardware-attested quote.
//
// Parameters:
//   - w: http.ResponseWriter used to send the HTTP response.
//   - r: *http.Request containing the incoming HTTP request.
//
// Workflow:
//  1. Parse and validate the request body.
//  2. Decode and validate the base64-encoded challenge.
//  3. Compute userData = SHA256(challenge || tlsFingerprint).
//  4. Fetch the attestation evidence using the computed userData.
//  5. Parse the evidence into a structured format.
//  6. Send the response with the attestation quote and certificate fingerprint.
func GetTdxQuote(w http.ResponseWriter, r *http.Request) {

	// ## REQUEST ##
	var reqBody types.GetTdxQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// ### BUSINESS LOGIC ###
	decodedChallenge, err := challenge.DecodeAndValidate(reqBody.Challenge)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	tlsFingerprint := tlscertificate.GetFingerprint()

	// Compute userData = SHA256(challenge || tlsFingerprint).
	// This binds both the nonce and certificate to the hardware quote.
	var userData [64]byte
	if tlsFingerprint != "" {
		combined := append(decodedChallenge[:], []byte(tlsFingerprint)...)
		hash := sha256.Sum256(combined)
		copy(userData[:32], hash[:])
		copy(userData[32:], decodedChallenge[:32])
	} else {
		userData = decodedChallenge
	}

	quoteStr, err := attestationquote.FetchEvidence(userData)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var quoteData map[string]any
	err = json.Unmarshal([]byte(quoteStr), &quoteData)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Failed to parse quote data")
		return
	}

	// ## RESPONSE ##
	resBody := types.GetTdxQuoteResponse{
		Status: "success",
		Data: types.AttestationQuoteEvidence{
			Quote:                     quoteData,
			TlsCertificateFingerprint: tlsFingerprint,
			UserDataComposition:       "SHA256(challenge || tlsFingerprint)[0:32] || challenge[0:32]",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}
