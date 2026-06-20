# Design

## Components

- **api/** — Go HTTP service. sqlite-backed flag store (`api/database`), a hand-rolled router, and an evaluation engine. Swagger docs are generated from annotations via `swag init`.
- **app/** — Static site (nginx-served) demonstrating flag consumption. On load it calls `/flags/dark-mode/evaluate` and toggles a `dark` CSS class based on the response.

## Data model

```go
type Flag struct {
    Key         string
    Description string
    IsEnabled   bool
    Rules       []Rule
    CreatedAt   time.Time
}

type Rule struct {
    Type      string   // "override" | "segment" | "percentage"
    UserIds   []string // override
    Attribute string   // segment
    Operator  string   // segment
    Value     string   // segment
    Rollout   int      // percentage, 0-99 inclusive bucket cutoff
}
```

A flag's `Rules` are evaluated **in order, first match wins**:

1. `override` — exact `userId` match always wins.
2. `segment` — context attribute equality check (e.g. `plan == "enterprise"`).
3. `percentage` — deterministic hash-based rollout.

If `IsEnabled` is `false`, evaluation short-circuits before any rule runs. If no rule matches, the flag evaluates to `false` with reason `no_rule_matched`.

## Sticky percentage rollout

```go
func getUserBucket(flagKey, userId string) int {
    input := fmt.Sprintf("%s:%s", flagKey, userId)
    sum := md5.Sum([]byte(input))
    num := binary.BigEndian.Uint32(sum[:4])
    return int(num % 100) // 0-99
}
```

A user lands in the rollout if `bucket < rule.Rollout`. Hashing `flagKey:userId` means the same user gets a consistent answer for a given flag across requests and servers, and raising the rollout percentage only ever adds users — it never evicts ones already in.

## Storage

Flags are persisted in sqlite (`github.com/mattn/go-sqlite3`), via the `api/database` package:

- `db.Init(path)` opens the database, creates the `flags` table if it doesn't exist, and seeds a `dark-mode` flag (`INSERT OR IGNORE`, so it's a no-op after the first run). The parent directory of `path` is created automatically if missing.
- `Rules` is stored as a JSON-encoded text column rather than a normalized table — rule shapes vary by type (`override`/`segment`/`percentage` each use a different subset of fields), so a flexible blob avoids a wide, mostly-NULL schema. `db.Flag`/`db.Rule` mirror the `main` package's `Flag`/`Rule` but live in `database` to avoid an import cycle (`main` depends on `database`, not the reverse); `fromDBFlag`/`toDBFlag` in `api/helpers.go` convert between them.
- `CreateFlag` does a plain `INSERT` and relies on the `key` primary key to reject duplicates; `UpdateFlag` is a full-row `UPDATE` keyed on `key`.
- No connection pooling concerns — sqlite with the default driver settings is fine at this scale, and the single `*sql.DB` handle is safe for concurrent use (the driver serializes writes internally).
- No migrations: schema changes mean dropping/recreating the dev database file. The seeded `dark-mode` row means a fresh `flags.db` is immediately useful without any manual setup.

There's no volume mount for the sqlite file in `docker-compose.yml`, so data does not survive a container recreate — same lifetime as the in-memory store this replaced. Add a volume on `/app/data` if persistence across restarts is needed.

## Partial updates (`PATCH /flags/{key}`)

The handler reads the existing flag from sqlite, overlays only the fields present in the request body, then writes the merged flag back with `db.UpdateFlag`. The request body is a dedicated `FlagPatch` type, not `Flag`:

```go
type FlagPatch struct {
    Description *string `json:"description"`
    IsEnabled   *bool   `json:"isEnabled"`
    Rules       []Rule  `json:"rules"`
}
```

`Description`/`IsEnabled` are pointers so the handler can tell "field omitted" apart from "field explicitly set to its zero value." A bare `bool` couldn't make that distinction — an earlier version reused `Flag` as the patch body, so any PATCH that omitted `isEnabled` silently reset it to `false`. `Rules` still replaces wholesale when present (no per-rule merging) — to change a single rollout percentage, send just that rule: `{"rules":[{"type":"percentage","rollout":100}]}`.

## API/frontend contract

`EvalResult` is the single source of truth for the evaluate response shape:

```go
type EvalResult struct {
    FlagKey   string `json:"flagKey"`
    IsEnabled bool   `json:"isEnabled"`
    Reason    string `json:"reason"`
}
```

The frontend must read `data.isEnabled` (not `data.enabled`) — a prior mismatch here caused the dark-mode flag to always appear off regardless of server-side evaluation.

## Why a hand-rolled router

The route set is small and stable (`/flags`, `/flags/{key}`, `/flags/{key}/evaluate`), so a `switch` on path patterns in `router.go` avoids pulling in a routing framework. Swagger annotations are bound per-function by `swag`, so each HTTP verb on a given path needs its own handler function (e.g. `handleListFlags` / `handleCreateFlag`) rather than one function branching on `r.Method` — otherwise only the last annotation block above the function gets picked up.
