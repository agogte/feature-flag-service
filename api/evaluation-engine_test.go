package main

import "testing"

func TestGetUserBucketIsStableAndInRange(t *testing.T) {
	for _, id := range []string{"a", "user_942", "", "ünïcode"} {
		first := getUserBucket("dark-mode", id)
		if first < 0 || first > 99 {
			t.Fatalf("bucket %d out of range for %q", first, id)
		}
		if again := getUserBucket("dark-mode", id); again != first {
			t.Fatalf("bucket not sticky for %q: %d then %d", id, first, again)
		}
	}
}

func TestEvaluate(t *testing.T) {
	enabled := func(rules ...Rule) Flag {
		return Flag{Key: "f", IsEnabled: true, Rules: rules}
	}

	tests := []struct {
		name       string
		flag       Flag
		userId     string
		ctx        map[string]string
		wantOn     bool
		wantReason string
	}{
		{"disabled flag ignores rules", Flag{Key: "f", Rules: []Rule{{Type: "percentage", Rollout: 100}}}, "u", nil, false, "flag_disabled"},
		{"no rules", enabled(), "u", nil, false, "no_rule_matched"},
		{"override match", enabled(Rule{Type: "override", UserIds: []string{"u1", "u2"}}), "u2", nil, true, "override"},
		{"override miss", enabled(Rule{Type: "override", UserIds: []string{"u1"}}), "u2", nil, false, "no_rule_matched"},
		{"segment match", enabled(Rule{Type: "segment", Attribute: "plan", Operator: "equals", Value: "enterprise"}), "u", map[string]string{"plan": "enterprise"}, true, "segment_match"},
		{"segment value differs", enabled(Rule{Type: "segment", Attribute: "plan", Operator: "equals", Value: "enterprise"}), "u", map[string]string{"plan": "free"}, false, "no_rule_matched"},
		{"segment attribute missing", enabled(Rule{Type: "segment", Attribute: "plan", Operator: "equals", Value: "enterprise"}), "u", nil, false, "no_rule_matched"},
		{"segment unknown operator", enabled(Rule{Type: "segment", Attribute: "plan", Operator: "contains", Value: "ent"}), "u", map[string]string{"plan": "enterprise"}, false, "no_rule_matched"},
		{"rollout 0%", enabled(Rule{Type: "percentage", Rollout: 0}), "u", nil, false, "no_rule_matched"},
		{"rollout 100%", enabled(Rule{Type: "percentage", Rollout: 100}), "u", nil, true, ""},
		{"later rule can still match", enabled(
			Rule{Type: "override", UserIds: []string{"other"}},
			Rule{Type: "segment", Attribute: "plan", Operator: "equals", Value: "pro"},
		), "u", map[string]string{"plan": "pro"}, true, "segment_match"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluate(tt.flag, tt.userId, tt.ctx)
			if got.IsEnabled != tt.wantOn {
				t.Errorf("IsEnabled = %v, want %v (reason %q)", got.IsEnabled, tt.wantOn, got.Reason)
			}
			if tt.wantReason != "" && got.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", got.Reason, tt.wantReason)
			}
			if got.FlagKey != tt.flag.Key {
				t.Errorf("FlagKey = %q, want %q", got.FlagKey, tt.flag.Key)
			}
		})
	}
}

func TestPercentageRolloutIsRoughlyProportional(t *testing.T) {
	flag := Flag{Key: "rollout", IsEnabled: true, Rules: []Rule{{Type: "percentage", Rollout: 30}}}
	const n = 10000
	on := 0
	for i := 0; i < n; i++ {
		if evaluate(flag, "user_"+itoa(i), nil).IsEnabled {
			on++
		}
	}
	if pct := float64(on) / n * 100; pct < 25 || pct > 35 {
		t.Fatalf("expected ~30%% enabled, got %.1f%%", pct)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for ; i > 0; i /= 10 {
		b = append([]byte{byte('0' + i%10)}, b...)
	}
	return string(b)
}
