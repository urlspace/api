package server

import (
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	response := &statusResponse{
		Status: "ok",
		Data:   "service is running",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
