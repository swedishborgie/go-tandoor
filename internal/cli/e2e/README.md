# CLI End-to-End Tests

These tests build the CLI binary and exercise entrypoints against a real Tandoor instance started via podman-compose.

## Run

```bash
INTEGRATION_TESTS=1 go test -tags=integration ./internal/cli/e2e -v
```

Single test:
```bash
INTEGRATION_TESTS=1 go test -tags=integration ./internal/cli/e2e -run TestE2ERecipesCreateAndGet -v
```

## What is tested

* CLI builds once per suite via `go build ./cmd/tandoor-cli`
* TestMain starts compose stack, creates superuser, obtains API token
* Tests run CLI via `exec.Command` with `TANDOOR_BASE_URL` and `TANDOOR_TOKEN`
* Priority 1 commands: `recipes list/create/get`, `foods list/ensure/get`, dry-run semantics, `--page-size`, JSON output

## Adding tests

Add `//go:build integration` file in this package, use `runCLI(args...)` helper.

`runCLI` returns stdout, stderr, exit code. CLI output is JSON for most commands.
