package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"evidence-verifier/internal/service/attestationquote"
	"evidence-verifier/internal/types"
	"evidence-verifier/internal/util"
	"github.com/go-playground/validator/v10"
)

// VerifyTdxQuote is an HTTP handler that accepts a JSON-encoded TDX quote,
// invokes the attestation verification service, and returns a JSON response.
//
// Expects a POST body matching VerifyTdxQuoteRequest (with a "quote" object).
// On parse errors, responds 400 Bad Request.
// On verification failures, responds 422 Unprocessable Entity with error details.
// On success, returns 200 OK with the VerificationResult in the response body.
func VerifyTdxQuote(w http.ResponseWriter, r *http.Request) {

	// ## REQUEST ##
	var reqBody types.VerifyTdxQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Validation error: "+err.Error()+". Please provide a valid tdx quote.")
		return
	}

	// ## BUSINESS LOGIC ##
	issuedChallenge := reqBody.IssuedChallenge
	manifestUrl := reqBody.BaselineManifestUrl
	tlsFingerprint := reqBody.TlsCertificateFingerprint
	evidenceQuote := reqBody.Quote

	verification, err := attestationquote.Verify(issuedChallenge, manifestUrl, tlsFingerprint, evidenceQuote)
	if err != nil {
		util.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// ## RESPONSE ##
	respBody := types.VerifyTdxQuoteResponse{
		Status: "success",
		Data:   verification,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respBody); err != nil {
		log.Printf("failed to write response: %v\n", err)
	}
}
