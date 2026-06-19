package main

import (
	"net/http"
	"time"
)

func handleFlags(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleListFlags(w, r)
	case http.MethodPost:
		handleCreateFlag(w, r)
	default:
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// GetFlags godoc
// @Summary     List all flags
// @Tags        flags
// @Produce     json
// @Success     200  {array}   Flag
// @Router      /flags [get]
func handleListFlags(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	flags := make([]Flag, 0, len(store))
	for _, f := range store {
		flags = append(flags, f)
	}
	mu.RUnlock()
	writeJson(w, http.StatusOK, flags)
}

// CreateFlag godoc
// @Summary     Create a flag
// @Tags        flags
// @Accept      json
// @Produce     json
// @Param       flag  body      Flag  true  "Flag to create"
// @Success     201   {object}  Flag
// @Failure     400   {object}  map[string]string
// @Router      /flags [post]
func handleCreateFlag(w http.ResponseWriter, r *http.Request) {
	var body Flag
	if err := readJson(r, &body); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "invalid Json"})
		return
	}
	if body.Key == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "key is required"})
		return
	}

	body.CreatedAt = time.Now()
	if body.Rules == nil {
		body.Rules = []Rule{}
	}

	mu.Lock()
	store[body.Key] = body
	mu.Unlock()
	writeJson(w, http.StatusCreated, body)
}

// EvaluateFlag godoc
// @Summary     Evaluate a flag for a user
// @Tags        flags
// @Produce     json
// @Param       key     path   string  true   "Flag key"
// @Param       userId  query  string  true   "User ID"
// @Param       plan    query  string  false  "User plan (e.g. enterprise)"
// @Success     200     {object}  EvalResult
// @Failure     400     {object}  map[string]string
// @Failure     404     {object}  map[string]string
// @Router      /flags/{key}/evaluate [get]
func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}

	key := flagKeyFromPath(r.URL.Path)
	userId := r.URL.Query().Get("userId")

	if userId == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"key": "userId is required"})
		return
	}

	mu.RLock()
	flag, ok := store[key]
	mu.RUnlock()

	if !ok {
		writeJson(w, http.StatusNotFound, map[string]string{"error": "flag not found"})
		return
	}

	context := map[string]string{}
	for k, vals := range r.URL.Query() {
		if k != "userId" {
			context[k] = vals[0]
		}
	}

	result := evaluate(flag, userId, context)
	writeJson(w, http.StatusOK, result)
}

// UpdateFlag godoc
// @Summary     Update a flag
// @Tags        flags
// @Accept      json
// @Produce     json
// @Param       key   path      string  true  "Flag key"
// @Param       flag  body      Flag    true  "Fields to update"
// @Success     200   {object}  Flag
// @Failure     404   {object}  map[string]string
// @Router      /flags/{key} [patch]
func handleUpdateFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	key := flagKeyFromPath(r.URL.Path)

	mu.RLock()
	existing, ok := store[key]
	mu.RUnlock()

	if !ok {
		writeJson(w, http.StatusNotFound, map[string]string{"error": "flag not found"})
		return
	}

	var body Flag
	if err := readJson(r, &body); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	// Only update fields that were sent
	if body.Rules != nil {
		existing.Rules = body.Rules
	}
	if body.Description != "" {
		existing.Description = body.Description
	}
	// Explicitly check IsEnabled since false is a valid value
	existing.IsEnabled = body.IsEnabled

	mu.Lock()
	store[key] = existing
	mu.Unlock()

	writeJson(w, http.StatusOK, existing)
}
