package render

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, body any, status int){
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		http.Error(w, "json encode error", http.StatusBadRequest)
		panic(err)
	}
}