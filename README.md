# Feature Flag Service

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/go-1.23-00ADD8?logo=go&logoColor=white">
  <img alt="SQLite" src="https://img.shields.io/badge/storage-SQLite-003B57?logo=sqlite&logoColor=white">
  <img alt="Docker" src="https://img.shields.io/badge/docker-compose-2496ED?logo=docker&logoColor=white">
  <img alt="OpenAPI" src="https://img.shields.io/badge/docs-Swagger-85EA2D?logo=swagger&logoColor=black">
  <img alt="Feature flags" src="https://img.shields.io/badge/feature%20flags-percentage%20rollouts-8A2BE2">
</p>

<p align="center"><b>English</b> · <a href="README.fr.md">Français</a></p>

A lightweight feature flag API with rule-based evaluation (override lists, segment targeting, and sticky percentage rollouts), plus a static docs site that demos it in light/dark mode via a live flag.

See [DESIGN.md](DESIGN.md) for architecture and evaluation-engine details.

## Screenshots

| Light mode (flag off) | Dark mode (flag on) |
| --- | --- |
| ![Light mode](documentation/light-mode.png) | ![Dark mode](documentation/dark-mode.png) |

Swagger UI:

![Swagger UI](documentation/swagger.png)

## Project layout

```
api/    Go API (sqlite-backed flag store, evaluation engine, swagger docs)
app/    Static docs site that reads the dark-mode flag and toggles its theme
docker-compose.yml
```

## Running it

```bash
docker-compose up --build
```

- API: http://localhost:3000 (Swagger UI at `/swagger/index.html`)
- App: http://localhost:8080

Flags are persisted to a sqlite file at `api/data/flags.db` (created automatically on first run). A `dark-mode` flag is seeded on startup if it doesn't already exist, so the API is usable immediately without any setup.

## API quick reference

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/flags` | List all flags |
| `POST` | `/flags` | Create a flag |
| `PATCH` | `/flags/{key}` | Update a flag's `isEnabled`/`rules`/`description` |
| `GET` | `/flags/{key}/evaluate?userId=...` | Evaluate a flag for a user |
| `GET` | `/health` | Liveness/readiness check (pings the database) |
| `GET` | `/metrics` | Flag count, eval count, error count, uptime |

Example — roll a flag out to 100% of users:

```bash
curl -X PATCH http://localhost:3000/flags/dark-mode \
  -H "Content-Type: application/json" \
  -d '{ "isEnabled": true, "rules": [{ "type": "percentage", "rollout": 100 }] }'

curl "http://localhost:3000/flags/dark-mode/evaluate?userId=user_942"
```

Ops endpoints:

```bash
curl http://localhost:3000/health
# { "status": "Ok", "database": true, "uptimeSeconds": 42 }

curl http://localhost:3000/metrics
# { "flagCount": 2, "evalCount": 17, "errorCount": 1, "uptimeSeconds": 42, "lastError": "...", "lastErrorAt": "..." }
```
