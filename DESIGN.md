# Design

## Components

- **api/** — Go HTTP service. In-memory flag store (`map[string]Flag` guarded by a `sync.RWMutex`), a hand-rolled router, and an evaluation engine. Swagger docs are generated from annotations via `swag init`.
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

## Concurrency

The store is a plain map protected by `mu sync.RWMutex`. Reads (`GET /flags`, evaluate) take `RLock`; writes (`POST /flags`, `PATCH /flags/{key}`) take `Lock`. There's no persistence — state resets on restart, with `dark-mode` re-seeded in `main.go`.

## Partial updates (`PATCH /flags/{key}`)

The handler reads the existing flag, then overlays only the fields present in the request body:

- `Rules` replaces wholesale if non-nil (no per-rule merging).
- `Description` replaces if non-empty.
- `IsEnabled` always overwrites — there's no way to distinguish "field omitted" from "explicitly set to `false`" with a plain `bool`, so a PATCH that omits `isEnabled` will reset it to `false`. Callers should always send `isEnabled` explicitly.

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
