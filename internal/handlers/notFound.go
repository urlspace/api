package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zapi-sh/api/internal/models"
)

func NotFound(w http.ResponseWriter, r *http.Request) {

	response := models.ResponseSuccess{
		Status: "error",
		Data:   "endpoint not found",
	}

	w.WriteHeader(http.StatusNotFound)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
