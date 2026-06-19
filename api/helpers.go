package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func writeJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func readJson(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func flagKeyFromPath(path string) string {
	//path: /flags/dark-mode or /flags/dark-mode/evaluate
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
