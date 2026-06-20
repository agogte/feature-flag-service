package db

import (
	"encoding/json"
	"time"
)

func GetFlag(key string) (Flag, bool) {
	row := DB.QueryRow(`SELECT key, description, isEnabled, rules, createdAt from flags
						WHERE key = ?`, key)

	var f Flag
	var rulesRaw string
	var createdAt string

	err := row.Scan(&f.Key, &f.Description, &f.IsEnabled, &rulesRaw, &createdAt)
	if err != nil {
		return Flag{}, false
	}

	json.Unmarshal([]byte(rulesRaw), &f.Rules)
	f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return f, true
}

func GetAllFlags() []Flag {
	rows, err := DB.Query(`SELECT key, description, isEnabled, rules, createdAt FROM flags`)
	if err != nil {
		return []Flag{}
	}
	defer rows.Close()

	flags := []Flag{}
	for rows.Next() {
		var f Flag
		var rulesRaw string
		var createdAt string

		rows.Scan(&f.Key, &f.Description, &f.IsEnabled, &rulesRaw, &createdAt)
		json.Unmarshal([]byte(rulesRaw), &f.Rules)
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		flags = append(flags, f)
	}

	return flags
}

func CreateFlag(f Flag) error {
	rules, _ := json.Marshal(f.Rules)

	_, err := DB.Exec(
		`INSERT INTO flags (key, description, isEnabled, rules, createdAt) VALUES (?, ?, ?, ?, ?)`,
		f.Key, f.Description, f.IsEnabled, string(rules), f.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func UpdateFlag(f Flag) error {
	rules, _ := json.Marshal(f.Rules)

	_, err := DB.Exec(
		`UPDATE flags SET description = ?, isEnabled = ?, rules = ? WHERE key = ?`,
		f.Description, f.IsEnabled, string(rules), f.Key,
	)
	return err
}
