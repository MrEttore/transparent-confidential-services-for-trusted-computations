package handler

import (
	"encoding/json"
	"evidence-provider/internal/service/challenge"
	"evidence-provider/internal/service/infrastructure"
	"evidence-provider/internal/types"
	"evidence-provider/internal/util"
	"net/http"
)

// GetInfrastructureSummary handles HTTP requests to fetch a summary of the infrastructure metadata.
// It validates the incoming request, decodes the challenge, fetches the infrastructure evidence,
// and returns the summary data in the response.
//
// Parameters:
//   - w: http.ResponseWriter used to send the HTTP response.
//   - r: *http.Request containing the incoming HTTP request.
//
// Workflow:
//  1. Parse and validate the request body.
//  2. Decode and validate the base64-encoded challenge.
//  3. Fetch the infrastructure evidence and bind it to the challenge.
//  4. Construct the response with the fetched evidence.
func GetInfrastructureSummary(w http.ResponseWriter, r *http.Request) {

	// ## REQUEST ##
	var reqBody types.GetInfrastructureSummaryRequest
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

	evidence, err := infrastructure.FetchEvidence(r.Context(), reqBody.Challenge)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// ## RESPONSE ##
	resBody := types.GetInfrastructureSummaryResponse{
		Status: "success",
		Data:   evidence,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}
