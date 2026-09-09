# Integration Tests

These tests spin up a fresh Tandoor instance via podman-compose, run migrations, create a superuser, obtain an API token, and run Go client tests against the real server.

## Run

```bash
# from repo root
INTEGRATION_TESTS=1 go test -tags=integration ./internal/tests -v
```

Or run a single test:
```bash
INTEGRATION_TESTS=1 go test -tags=integration ./internal/tests -run TestIntegrationRecipeBookCRUD -v
```

Environment:
* `podman-compose` or the `docker compose` v2 plugin must be on PATH (podman-compose is preferred)
* port 8080 must be free
* Docker images will be pulled: ghcr.io/tandoorrecipes/recipes:2.6.13, postgres:16-alpine

Tests are ephemeral: `podman-compose down -v` removes DB and static volumes on exit.

## Files

* `docker-compose.test.yml` – test compose
* `.env.test.tmpl` – template for env vars
* `compose.go` – start/stop compose
* `env.go` – generate random SECRET_KEY / POSTGRES_PASSWORD
* `auth.go` – create superuser + fetch token
* `suite_test.go` – TestMain, wait for readiness, expose `Client()`
