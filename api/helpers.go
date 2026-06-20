package main

import (
	"encoding/json"
	"net/http"
	"strings"

	db "github.com/agogte/feature-flag-service/api/database"
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

func fromDBFlag(f db.Flag) Flag {
	rules := make([]Rule, len(f.Rules))
	for i, r := range f.Rules {
		rules[i] = Rule{
			Type:      r.Type,
			UserIds:   r.UserIDs,
			Attribute: r.Attribute,
			Operator:  r.Operator,
			Value:     r.Value,
			Rollout:   r.Rollout,
		}
	}
	return Flag{
		Key:         f.Key,
		Description: f.Description,
		IsEnabled:   f.IsEnabled,
		Rules:       rules,
		CreatedAt:   f.CreatedAt,
	}
}

func toDBFlag(f Flag) db.Flag {
	rules := make([]db.Rule, len(f.Rules))
	for i, r := range f.Rules {
		rules[i] = db.Rule{
			Type:      r.Type,
			UserIDs:   r.UserIds,
			Attribute: r.Attribute,
			Operator:  r.Operator,
			Value:     r.Value,
			Rollout:   r.Rollout,
		}
	}
	return db.Flag{
		Key:         f.Key,
		Description: f.Description,
		IsEnabled:   f.IsEnabled,
		Rules:       rules,
		CreatedAt:   f.CreatedAt,
	}
}
