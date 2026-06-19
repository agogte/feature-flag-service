package main

import (
	"sync"
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
	store = map[string]Flag{}
	mu    sync.RWMutex
)
