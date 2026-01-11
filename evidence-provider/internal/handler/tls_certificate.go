package handler

import (
	"encoding/json"
	"net/http"

	"evidence-provider/internal/service/challenge"
	"evidence-provider/internal/service/tlscertificate"
	"evidence-provider/internal/types"
	"evidence-provider/internal/util"
)

// GetTlsCertificate handles HTTP requests to fetch TLS certificate evidence.
// It returns the certificate fingerprint bound to the client's challenge nonce.
//
// Workflow:
//  1. Parse and validate the request body.
//  2. Decode and validate the base64-encoded challenge.
//  3. Fetch the certificate evidence and bind it to the challenge.
//  4. Construct the response with the fetched evidence.
func GetTlsCertificate(w http.ResponseWriter, r *http.Request) {

	// ## REQUEST ##
	var reqBody types.GetTlsCertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// ## BUSINESS LOGIC ##
	_, err := challenge.DecodeAndValidate(reqBody.Challenge)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	evidence, err := tlscertificate.FetchEvidence(r.Context(), reqBody.Challenge)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// ## RESPONSE ##
	resBody := types.GetTlsCertificateResponse{
		Status: "success",
		Data:   evidence,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}
