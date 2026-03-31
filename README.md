# Patwos API

Go REST API with JWT auth, articles, comments, voting, built for my site (but i suppose could be used elsewhere)

## architecture

```
request → gin router → middleware chain → controller → service → repository → postgres
```

layered with interfaces at every boundary (repository pattern, service pattern), every dependency is injected, nothing is hardcoded, every layer is independently testable with fakes

## database migrations

versioned sql migrations via [golang-migrate](https://github.com/golang-migrate/migrate), embedded in the binary with `//go:embed`. no auto-migrations in production.

```bash
# apply pending migrations
./main migrate

# fix dirty state (drops schema_migrations table, re-runs from scratch)
./main migrate-force 0

# force to specific version (recovery only)
./main migrate-force 3

# start the server
./main serve
```

## docker deployment

```bash
# run migrations first (deliberate, separate step)
docker compose run --rm patwos-api ./main migrate

# then start the stack
docker compose up -d
```

## middleware ordering

```
recovery → security headers → cors → rate limit → timeout → body limit → logger
```

**why this order matters**: recovery must be first to catch panics in any middleware. security headers go before cors so every response gets them. cors is before rate limit so preflight OPTIONS requests don't burn rate limit tokens. body limit is after timeout so oversized requests don't bypass the deadline. logger is last so it captures the final status code after all middleware has run.

## configuration

all config via environment variables with sensible defaults, required secrets (`JWT_SECRET`, `DB_PASSWORD`) fail hard at startup with `log.Fatal` instead of falling back to insecure defaults.

| variable | default | purpose |
|----------|---------|---------|
| `DB_HOST` | `localhost` | postgres host |
| `DB_PORT` | `5432` | postgres port |
| `DB_SSLMODE` | `disable` | tls mode (warns in release mode if disabled) |
| `API_PORT` | `8080` | http listen port |
| `GIN_MODE` | `debug` | gin framework mode |
| `ALLOWED_ORIGINS` | `*` | cors origins (comma-separated) |
| `MAX_REQUEST_SIZE` | `10MB` | body size limit |
| `REQUEST_TIMEOUT` | `15s` | per-request deadline |
| `READ_TIMEOUT` | `10s` | http read timeout |
| `WRITE_TIMEOUT` | `15s` | http write timeout |
| `IDLE_TIMEOUT` | `120s` | keep-alive timeout |
| `SHUTDOWN_TIMEOUT` | `30s` | graceful shutdown deadline |

## graceful shutdown

the server catches `SIGINT`/`SIGTERM`, stops accepting new connections, drains in-flight requests within the shutdown timeout, then exits cleanly. docker sends `SIGTERM` on `docker stop`, so containers shutdown gracefully without dropped requests.

## project structure

```
main.go                 → entrypoint with subcommands (serve, migrate, migrate-force)
config/                 → env-based configuration loading
database/               → connection pool setup, migration runner
middleware/             → auth, cors, rate limit, security headers, body limit, timeout, logger, admin
controllers/            → http handlers (thin, delegate to services)
service/                → business logic interfaces + implementations
repository/             → data access interfaces + gorm implementations
models/                 → domain types, request/response dtos, validation tags
authcache/              → in-memory revoked token cache with ttl cleanup
migrations/             → versioned sql migration files (embedded in binary)
.github/workflows/      → ci/cd pipelines
```

## author

Wosiu6
