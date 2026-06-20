package main

import (
	"time"
)

// Flag represents a feature flag and its targeting rules.
type Flag struct {
	Key         string    `json:"key"`
	Description string    `json:"description"`
	IsEnabled   bool      `json:"isEnabled"`
	Rules       []Rule    `json:"rules"`
	CreatedAt   time.Time `json:"createdAt"`
}

// FlagPatch represents a partial update to a flag. Pointer fields let the
// handler tell "field omitted" apart from "field explicitly set to its zero
// value" (e.g. isEnabled: false).
type FlagPatch struct {
	Description *string `json:"description"`
	IsEnabled   *bool   `json:"isEnabled"`
	Rules       []Rule  `json:"rules"`
}

// EvalResult is the response from the evaluate endpoint.
type EvalResult struct {
	FlagKey   string `json:"flagKey"`
	IsEnabled bool   `json:"isEnabled"`
	Reason    string `json:"reason"`
}

// Rule is a single targeting rule inside a flag.
type Rule struct {
	Type      string   `json:"type"`
	UserIds   []string `json:"userIds"`
	Attribute string   `json:"attribute"`
	Operator  string   `json:"operator"`
	Value     string   `json:"value"`
	Rollout   int      `json:"rollout"`
}

var (
	evalCount   int64     // total evaluate calls since boot
	errorCount  int64     // total failed requests since boot
	lastError   string    // most recent error message
	lastErrorAt time.Time // when it happened
	startedAt   = time.Now()
)
