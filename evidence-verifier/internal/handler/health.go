package handler

import (
	"encoding/json"
	"evidence-verifier/internal/types"
	"evidence-verifier/internal/util"
	"net/http"
)

// Health checks if the Evidence Verifier is up and running.
func Health(w http.ResponseWriter, r *http.Request) {
	res := types.HealthResponse{
		Message: "Evidence Verifier is healthy",
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}
