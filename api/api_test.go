package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	db "github.com/agogte/feature-flag-service/api/database"
)

func setup(t *testing.T) {
	t.Helper()
	db.Init(filepath.Join(t.TempDir(), "flags.db"))
	t.Cleanup(func() { db.DB.Close() })
}

func do(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	router(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestHealth(t *testing.T) {
	setup(t)
	rec := do(t, http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := decode[map[string]any](t, rec)
	if body["database"] != true {
		t.Fatalf("database = %v", body["database"])
	}
}

func TestCreateListAndConflict(t *testing.T) {
	setup(t)

	if rec := do(t, http.MethodPost, "/flags", `{"key":"new-flag","isEnabled":true}`); rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d: %s", rec.Code, rec.Body)
	}
	if rec := do(t, http.MethodPost, "/flags", `{"key":"new-flag"}`); rec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d", rec.Code)
	}
	if rec := do(t, http.MethodPost, "/flags", `{"description":"no key"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing key status = %d", rec.Code)
	}
	if rec := do(t, http.MethodPost, "/flags", `not json`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json status = %d", rec.Code)
	}

	flags := decode[[]Flag](t, do(t, http.MethodGet, "/flags", ""))
	found := false
	for _, f := range flags {
		if f.Key == "new-flag" {
			found = true
		}
	}
	if !found {
		t.Fatalf("new-flag not in list: %+v", flags)
	}
}

func TestPatchOnlyChangesProvidedFields(t *testing.T) {
	setup(t)
	do(t, http.MethodPost, "/flags", `{"key":"p","description":"keep me","isEnabled":true}`)

	rec := do(t, http.MethodPatch, "/flags/p", `{"isEnabled":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	got := decode[Flag](t, rec)
	if got.IsEnabled || got.Description != "keep me" {
		t.Fatalf("unexpected flag after patch: %+v", got)
	}

	if rec := do(t, http.MethodPatch, "/flags/missing", `{}`); rec.Code != http.StatusNotFound {
		t.Fatalf("missing flag status = %d", rec.Code)
	}
	if rec := do(t, http.MethodGet, "/flags/p", ""); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET on patch route status = %d", rec.Code)
	}
}

func TestEvaluateEndpoint(t *testing.T) {
	setup(t)
	do(t, http.MethodPost, "/flags", `{"key":"e","isEnabled":true,"rules":[{"type":"segment","attribute":"plan","operator":"equals","value":"pro"}]}`)

	rec := do(t, http.MethodGet, "/flags/e/evaluate?userId=u1&plan=pro", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if res := decode[EvalResult](t, rec); !res.IsEnabled || res.Reason != "segment_match" {
		t.Fatalf("unexpected result: %+v", res)
	}

	if res := decode[EvalResult](t, do(t, http.MethodGet, "/flags/e/evaluate?userId=u1&plan=free", "")); res.IsEnabled {
		t.Fatalf("free plan should not match: %+v", res)
	}
	if rec := do(t, http.MethodGet, "/flags/e/evaluate", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing userId status = %d", rec.Code)
	}
	if rec := do(t, http.MethodGet, "/flags/nope/evaluate?userId=u1", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown flag status = %d", rec.Code)
	}
}

func TestMetricsCountsEvalsAndErrors(t *testing.T) {
	setup(t)
	before := decode[map[string]any](t, do(t, http.MethodGet, "/metrics", ""))

	do(t, http.MethodGet, "/flags/nope/evaluate?userId=u", "")

	after := decode[map[string]any](t, do(t, http.MethodGet, "/metrics", ""))
	if after["evalCount"].(float64) != before["evalCount"].(float64)+1 {
		t.Fatalf("evalCount did not increase: %v -> %v", before["evalCount"], after["evalCount"])
	}
	if after["errorCount"].(float64) != before["errorCount"].(float64)+1 {
		t.Fatalf("errorCount did not increase: %v -> %v", before["errorCount"], after["errorCount"])
	}
	if after["lastError"] == nil {
		t.Fatal("lastError missing")
	}
}

func TestUnknownRouteAndCORS(t *testing.T) {
	setup(t)
	if rec := do(t, http.MethodGet, "/nope", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	rec := do(t, http.MethodOptions, "/flags", "")
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("preflight: code=%d headers=%v", rec.Code, rec.Header())
	}
}
