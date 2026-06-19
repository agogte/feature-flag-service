package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

func getUserBucket(flagKey, userId string) int {
	input := fmt.Sprintf("%s:%s", flagKey, userId)
	sum := md5.Sum([]byte(input))
	num := binary.BigEndian.Uint32(sum[:4])
	return int(num % 100)
}

func evaluate(flag Flag, userId string, ctx map[string]string) EvalResult {
	if !flag.IsEnabled {
		return EvalResult{flag.Key, false, "flag_disabled"}
	}

	for _, rule := range flag.Rules {
		switch rule.Type {

		case "override":
			for _, id := range rule.UserIds {
				if id == userId {
					return EvalResult{flag.Key, true, "override"}
				}
			}
		case "segment":
			if val, ok := ctx[rule.Attribute]; ok {
				if rule.Operator == "equals" && val == rule.Value {
					return EvalResult{flag.Key, true, "segment_match"}
				}
			}

		case "percentage":
			bucket := getUserBucket(flag.Key, userId)
			if bucket < rule.Rollout {
				return EvalResult{flag.Key, true, "precentage_rollout"}
			}
		}
	}

	return EvalResult{flag.Key, false, "no_rule_matched"}
}
