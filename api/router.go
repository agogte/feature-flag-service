package main

import (
	"net/http"
	"strings"
)

func router(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

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
	case path == "/health":
		handleHealth(w, r)
	case path == "/metrics":
		handleMetrics(w, r)
	default:
		writeJson(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}
}
