package api

import (
	"encoding/json"
	"net/http"
)

func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	if _, err = w.Write(dat); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
