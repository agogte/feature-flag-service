package main

import (
	"net/http"
	"strings"
)

func router(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := r.URL.Path

	switch {
	case path == "/flags":
		handleFlags(w, r)
	case strings.HasSuffix(path, "/evaluate"):
		handleEvaluate(w, r)
	case strings.HasPrefix(path, "/flags/") && !strings.HasSuffix(path, "/evaluate"):
		handleUpdateFlag(w, r)
	default:
		writeJson(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}
}
