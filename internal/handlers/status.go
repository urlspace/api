package handlers

import (
	"encoding/json"
	"net/http"
)

type StatusResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func Status(w http.ResponseWriter, r *http.Request) {
	response := &StatusResponse{
		Status: "ok",
		Data:   "service is running",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
