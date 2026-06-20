package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Flag and Rule are duplicated here (not imported from main) because
// the db package must not depend on main — main depends on db, not
// the other way around. This avoids an import cycle.
type Rule struct {
	Type      string   `json:"type"`
	UserIDs   []string `json:"userIds,omitempty"`
	Attribute string   `json:"attribute,omitempty"`
	Operator  string   `json:"operator,omitempty"`
	Value     string   `json:"value,omitempty"`
	Rollout   int      `json:"rollout,omitempty"`
}

type Flag struct {
	Key         string    `json:"key"`
	Description string    `json:"description"`
	IsEnabled   bool      `json:"isEnabled"`
	Rules       []Rule    `json:"rules"`
	CreatedAt   time.Time `json:"createdAt"`
}

func Init(path string) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal("failed to create db directory: ", err)
		}
	}

	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal("failed to open db: ", err)
	}

	createTable := `
		CREATE TABLE IF NOT EXISTS flags (
		key TEXT primary key,
		description text,
		isEnabled boolean not null default 0,
		rules text not null default '[]',
		createdAt text not null);
	`

	if _, err := DB.Exec(createTable); err != nil {
		log.Fatal("failed to create table: ", err)
	}

	seed()
}

func seed() {
	insert :=
		`INSERT OR IGNORE INTO FLAGS (key, description, isEnabled, rules, createdAt)
	values (?,?,?,?,?)`

	rules, _ := json.Marshal([]Rule{
		{Type: "percentage", Rollout: 50},
	})

	_, err := DB.Exec(insert, "dark-mode", "Switches UI to dark mode", true, string(rules), time.Now().Format(time.RFC3339))

	if err != nil {
		log.Println("seed error: ", err)
	}
}
