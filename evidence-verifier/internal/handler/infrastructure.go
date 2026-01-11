package handler

import (
	"encoding/json"
	"evidence-verifier/internal/service/infrastructure"
	"evidence-verifier/internal/types"
	"evidence-verifier/internal/util"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
)

func VerifyInfrastructure(w http.ResponseWriter, r *http.Request) {

	// ## REQUEST ##
	var reqBody types.VerifyInfrastructureRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(reqBody); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Validation error: "+err.Error()+". Please provide a valid infrastructure evidence, baseline manifest url, and challenge.")
		return
	}

	// ## BUSINESS LOGIC ##
	issuedChallenge := reqBody.IssuedChallenge
	manifestUrl := reqBody.BaselineManifestUrl
	infrastructureEvidence := reqBody.Evidence

	verification, err := infrastructure.Verify(issuedChallenge, manifestUrl, infrastructureEvidence)
	if err != nil {
		util.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// ## RESPONSE ##
	respBody := types.VerifyInfrastructureResponse{
		Status: "success",
		Data:   verification,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respBody); err != nil {
		log.Printf("failed to write response: %v\n", err)
	}

}
