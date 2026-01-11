package util

import (
	"encoding/json"
	"evidence-verifier/internal/types"
	"log"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(types.ErrorResponse{
		Status:  "error",
		Message: message,
	}); err != nil {
		log.Printf("failed to write response: %v\n", err)
	}
}
