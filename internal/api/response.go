package api

import (
	"fmt"
	"net/http"
	"strconv"

	json "github.com/json-iterator/go"
)

func respondJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: fix this mess
	if body, ok := data.([]byte); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(code)
		if _, err := w.Write(body); err != nil {
			http.Error(w, fmt.Sprintf("failed to write response body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	} else {

		j, err := json.Marshal(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal JSON response: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(j)))
		w.WriteHeader(code)
		if _, err = w.Write(j); err != nil {
			http.Error(w, fmt.Sprintf("failed to write response body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}
