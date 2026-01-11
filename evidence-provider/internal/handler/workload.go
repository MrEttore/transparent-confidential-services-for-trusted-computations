package handler

import (
	"context"
	"encoding/json"
	"evidence-provider/internal/service/challenge"
	"evidence-provider/internal/service/workload"
	"evidence-provider/internal/types"
	"evidence-provider/internal/util"
	"net/http"
)

// GetWorkloads handles HTTP requests to fetch metadata about running workloads inside the TEE.
// It validates the incoming request, decodes the challenge, fetches workload evidence,
// and returns the metadata in the response.
//
// Parameters:
//   - w: http.ResponseWriter used to send the HTTP response.
//   - r: *http.Request containing the incoming HTTP request.
//
// Workflow:
//  1. Parse and validate the request body.
//  2. Decode and validate the base64-encoded challenge.
//  3. Fetch the workload evidence and bind it to the challenge.
//  4. Construct the response with the fetched evidence.
func GetWorkloads(w http.ResponseWriter, r *http.Request) {

	// ## REQUEST ##
	var reqBody types.GetWorkloadsRequest
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

	evidence, err := workload.FetchEvidence(context.Background(), reqBody.Challenge)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// ## RESPONSE ##
	resBody := types.GetWorkloadsResponse{
		Status: "success",
		Data:   evidence,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}
