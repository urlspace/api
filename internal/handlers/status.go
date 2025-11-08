package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/zapi-sh/api/internal/models"
)

func Status(w http.ResponseWriter, r *http.Request) {
	time.Sleep(6 * time.Second)

	response := models.ResponseSuccess{
		Status: "success",
		Data:   "Service is running",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
