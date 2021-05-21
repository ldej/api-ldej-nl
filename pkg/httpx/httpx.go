package httpx

import (
	"encoding/json"
	"net/http"
)

// swagger:model ErrorResponse
type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, r *http.Request, code int, body interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	if _, ok := r.URL.Query()["pretty"]; ok {
		enc.SetIndent("", "  ")
	}

	if err := enc.Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func JSON(w http.ResponseWriter, r *http.Request, body interface{}) {
	writeJSON(w, r, http.StatusOK, body)
}

func AbortJSON(w http.ResponseWriter, r *http.Request, code int, err error) {
	writeJSON(w, r, code, ErrorResponse{
		Error: err.Error(),
	})
}
