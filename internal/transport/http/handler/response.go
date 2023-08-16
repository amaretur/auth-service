package handler

import (
	"net/http"
	"encoding/json"
)

func Response(w http.ResponseWriter, data interface{}) {

	if err := json.NewEncoder(w).Encode(&data); err != nil {

		Error(
			w, 
			http.StatusInternalServerError,
			"data acquisition error...",
		)
	}
}

func Error(w http.ResponseWriter, statusCode int, message string) {

	w.WriteHeader(statusCode)

	if message != "" {
		json.NewEncoder(w).Encode(map[string]string{
			"error": message,
		});
	}
}
