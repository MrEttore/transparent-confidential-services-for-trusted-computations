package handler

import (
	"encoding/json"
	"evidence-provider/internal/types"
	"evidence-provider/internal/util"
	"net/http"
)

// Health handles HTTP requests to check the health status of the Evidence Provider service.
// It responds with a JSON message indicating the service health status.
func Health(w http.ResponseWriter, r *http.Request) {
	res := types.HealthResponse{
		Message: "Evidence Provider is healthy",
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}
