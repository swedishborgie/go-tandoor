# MCP End-to-End Tests

These tests build the `tandoor-cli` binary and drive its `mcp` subcommand as
an MCP server over stdio against a real Tandoor instance (2.6.13) started
via podman-compose.

## Run

```bash
INTEGRATION_TESTS=1 go test -tags=integration ./internal/mcp/e2e -v
```

Single test:
```bash
INTEGRATION_TESTS=1 go test -tags=integration ./internal/mcp/e2e -run TestE2ERecipeWriteCycle -v
```

## What is tested

* MCP binary builds once per suite via `go build ./cmd/tandoor-cli`
* TestMain starts the compose stack on port 8081, creates a superuser, seeds
  space/household/meal-type/book data, and obtains an API token
* Each test spawns `tandoor-cli mcp` with `TANDOOR_BASE_URL`/`TANDOOR_TOKEN`
  and talks MCP over stdio (mcp-go stdio client)
* Write-tool lifecycle per resource: apply → patch/update → delete → verify
  gone (write tools perform the operation and return the result object)
* M3 composites: `property_attach` idempotency (create → update),
  `food_ensure` cycle (not_found → created → found), `food_audit_inspect`
  + `food_audit_fix` rename, `food_find_duplicates` grouping,
  `shopping_recipe_create_entries` scaled bulk create, `meal_plan_auto_plan`
  real plan, share-link create, access-token cycle, household cycle,
  sync-config cycle, `fdc_*` clean error without `FDC_API_KEY`
* Server modes: `TANDOOR_MCP_READ_ONLY` hides write tools, `TANDOOR_MCP_TOOLS`
  filters the catalog

## API quirks the tests work around

* Inventory entries require a `unit_id` (the API rejects `unit: null`).
* Standalone steps/ingredients are invisible to the API until attached to a
  visible recipe, so their write cycles create the objects inside a recipe
  payload first.
* Meal plans need a `to_date` within the visibility window (`now-90d`…`now+360d`)
  or they 404 on read — auto-plan tests therefore plan for tomorrow.
* Auto-plan only considers `internal=true` recipes.
* Fresh spaces have no units or property types (Tandoor seeds only
  pre-existing spaces), so `property_attach` tests pass `per_100_unit_id`.
* Steps require the `ingredients` key even when empty (`[]`).
* `all=true` list tools return a flat array (not the paginated envelope).

## Adding tests

Add a `//go:build integration` file in this package. Helpers:

* `newMCPClient(t)` — initialized stdio MCP client (full tool set)
* `newMCPClientWithEnv(t, "KEY=VAL", ...)` — client with extra env
* `callTool(t, c, name, args)` — call a tool, fail on tool error, return the
  parsed JSON object
* `callToolRaw(t, c, name, args)` — return raw text + error flag
* `runSuffix()` — unique suffix for object names
