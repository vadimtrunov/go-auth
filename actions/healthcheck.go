package actions

import (
	"encoding/json"
	"net/http"
)

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "Ok",
		"message": "I'm OK",
		"code":    200,
	}

	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
