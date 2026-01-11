package handler

import (
	"encoding/json"
	"evidence-verifier/internal/service/workload"
	"evidence-verifier/internal/types"
	"evidence-verifier/internal/util"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
)

func VerifyWorkloads(w http.ResponseWriter, r *http.Request) {

	// ## REQUEST ##
	var reqBody types.VerifyWorkloadsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Validation error: "+err.Error()+". Please provide a valid workload evidence, reference image, and challenge.")
		return
	}

	// ## BUSINESS LOGIC ##
	issuedChallenge := reqBody.IssuedChallenge
	referenceImage := reqBody.ReferenceImage
	evidence := reqBody.Evidence

	// Fetch reference metadata from Docker Hub’s Hub API and Registry HTTP API V2.
	workloadReferenceMetadata, err := workload.FetchWorkloadReferenceMetadata(r.Context(), &referenceImage)
	if err != nil {
		util.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// Verify evidence against golden reference values.
	verification, err := workload.Verify(issuedChallenge, &evidence, &referenceImage, &workloadReferenceMetadata)
	if err != nil {
		util.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// ## RESPONSE ##
	respBody := types.VerifyWorkloadsResponse{
		Status: "success",
		Data:   verification,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respBody); err != nil {
		log.Printf("failed to write response: %v\n", err)
	}
}
