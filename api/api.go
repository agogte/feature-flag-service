package main

import (
	"net/http"
	"time"

	db "github.com/agogte/feature-flag-service/api/database"
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
	dbFlags := db.GetAllFlags()
	flags := make([]Flag, 0, len(dbFlags))
	for _, f := range dbFlags {
		flags = append(flags, fromDBFlag(f))
	}

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

	if err := db.CreateFlag(toDBFlag(body)); err != nil {
		writeJson(w, http.StatusConflict, map[string]string{"error": "flag already exists"})
		return
	}

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
		return
	}

	key := flagKeyFromPath(r.URL.Path)
	userId := r.URL.Query().Get("userId")

	if userId == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"key": "userId is required"})
		return
	}

	dbFlag, ok := db.GetFlag(key)
	if !ok {
		writeJson(w, http.StatusNotFound, map[string]string{"error": "flag not found"})
		return
	}

	flag := fromDBFlag(dbFlag)

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
// @Param       key   path      string     true  "Flag key"
// @Param       flag  body      FlagPatch  true  "Fields to update"
// @Success     200   {object}  Flag
// @Failure     404   {object}  map[string]string
// @Router      /flags/{key} [patch]
func handleUpdateFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	key := flagKeyFromPath(r.URL.Path)

	dbFlag, ok := db.GetFlag(key)

	if !ok {
		writeJson(w, http.StatusNotFound, map[string]string{"error": "flag not found"})
		return
	}

	existing := fromDBFlag(dbFlag)

	var body FlagPatch
	if err := readJson(r, &body); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	// Only update fields that were actually present in the request
	if body.Rules != nil {
		existing.Rules = body.Rules
	}
	if body.Description != nil {
		existing.Description = *body.Description
	}
	if body.IsEnabled != nil {
		existing.IsEnabled = *body.IsEnabled
	}

	if err := db.UpdateFlag(toDBFlag(existing)); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "update flag failed"})
		return
	}

	writeJson(w, http.StatusOK, existing)
}
