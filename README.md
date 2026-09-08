# go-tandoor

A Go client library for the [Tandoor Recipes](https://tandoorRecipes.com/) REST API, plus
`tandoor-cli`, a full-featured command-line client for managing recipes, shopping lists,
meal plans, and food data — built for humans *and* AI agents.

- **Library** (`github.com/swedishborgie/go-tandoor`): typed, paginated access to the Tandoor
  REST API, covering recipes, ingredients, steps, foods, units & conversions, shopping lists,
  meal plans, cook logs, recipe books, properties, spaces, imports, and more.
- **CLI** (`tandoor-cli`): a single binary that wraps the API with sensible subcommands,
  batch operations, JSON output (`--jq` filters, `--output-file`), dry-run mode for writes,
  food-data audit tooling, and a client for the USDA FoodData Central (FDC) API.

## Installation

### CLI

```sh
go install github.com/swedishborgie/go-tandoor/cmd/tandoor-cli@latest
```

### Library

```sh
go get github.com/swedishborgie/go-tandoor
```

Requires Go 1.26+.

## Quickstart

### 1. Get an access token

Tokens can be created in Tandoor's web UI (Settings → API tokens, `tda_...`), or obtained
via the CLI:

```sh
# Username/password login (POST /api-token-auth/)
tandoor-cli --base-url https://recipes.example.com auth token --username you --password secret

# OIDC browser login (opens your browser, catches the callback on localhost:9999)
tandoor-cli --base-url https://recipes.example.com auth oidc --oidc-backend github
```

Both print the token to stdout. The CLI also has `auth list-tokens` for enumerating existing tokens.

### 2. Make requests

```sh
export TANDOOR_TOKEN=tda_...
export TANDOOR_BASE_URL=https://recipes.example.com   # required: your Tandoor instance URL

tandoor-cli recipes list --search "chicken"
tandoor-cli recipes get 42
tandoor-cli foods ensure --name "butternut squash" --name "thyme"
tandoor-cli shopping add-recipe 42 --list-id 1
```

Output is pretty-printed JSON. List commands support `--all` to collect every page into one
array, and many support `--jq '<filter>'` (built-in [gojq](https://github.com/itchyny/gojq), no
external `jq` binary needed) and `--output-file`.

Write commands support `--dry-run` to preview the exact payload without sending it.

## CLI reference

```
tandoor-cli [global options] <command> [command options]
```

### Global flags

| Flag | Alias | Env var | Description |
| --- | --- | --- | --- |
| `--base-url` | `-b` | `TANDOOR_BASE_URL` | Tandoor instance base URL (**required**) |
| `--token` | `-t` | `TANDOOR_TOKEN` | API access token (`tda_...`) |
| `--har-file` | | `TANDOOR_HAR_FILE` | Record all requests/responses to a HAR file (via [daytripper](https://github.com/swedishborgie/daytripper)) |
| `--page-size` | | | Page size for list operations (default `50`) |

FDC commands additionally accept `--fdc-api-key` (`FDC_API_KEY` env var) and honor
`FDC_BASE_URL` for the USDA FoodData Central API.

### Commands

| Command | Subcommands |
| --- | --- |
| `auth` | `token`, `oidc`, `list-tokens` |
| `recipes` | `list`, `get`, `create`, `update`, `patch`, `delete`, `batch-update`, `set-image`, `related`, `overview` |
| `shopping` | `list`, `get`, `create`, `delete`, `entries`, `add-entry`, `bulk-update`, `add-recipe`, `create-entries-from-recipe` |
| `mealplans` | `list`, `get`, `create`, `update`, `delete`, `ical`, `auto-plan` |
| `mealtypes` | `list`, `get` |
| `cook-logs` | `list`, `create` |
| `books` | `list`, `get`, `create`, `update`, `delete` |
| `book-entries` | `list`, `create`, `delete` |
| `imports` | `list`, `get`, `import`, `import-all`, `delete` |
| `import-logs` | `list` |
| `recipe-from-source` | `create` (import a recipe from a URL or raw payload) |
| `share-links` | `list`, `create`, `delete` |
| `foods` | `list`, `get`, `create`, `update`, `patch`, `delete`, `merge`, `ensure`, `attach-fdc-properties`, `auto-conversions` |
| `keywords` | `list`, `get`, `create`, `update`, `patch`, `delete`, `merge` |
| `units` | `list`, `get`, `create`, `update`, `patch`, `delete`, `merge` |
| `conversions` | `list`, `get`, `create`, `create-batch`, `update`, `patch`, `delete` |
| `ingredients` | `list`, `get`, `create`, `update`, `patch`, `delete` |
| `steps` | `list`, `get`, `create`, `update`, `patch`, `delete` |
| `properties` | `list`, `get`, `create`, `update`, `delete`, `attach` |
| `property-types` | `list`, `get`, `create`, `update`, `delete` |
| `audit` | `foods`, `food` (`inspect`, `suggest`, `fix`), `duplicates`, `connectors` |
| `fdc` | `search`, `food` (FDC ID lookup for a Tandoor food), `get` |

Run `tandoor-cli <command> --help` (or any subcommand) for flags and usage.

### Food-data audit

The `audit` group scans your food database for hygiene issues and can propose/apply
corrections, optionally cross-checking against USDA FoodData Central:

```sh
tandoor-cli audit foods --category quantity_prefix   # scan for naming issues
tandoor-cli audit duplicates --threshold 0.9         # find near-duplicate foods
tandoor-cli audit food inspect 123                   # deep-dive one food (needs FDC_API_KEY)
tandoor-cli audit food suggest 123                   # propose a correction
tandoor-cli audit food fix 123 --yes                 # apply it (--dry-run to preview)
tandoor-cli audit connectors                         # analyze +/-/& ingredient lines
```

## Library usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
)

func main() {
	client, err := tandoor.NewClient(
		"https://recipes.example.com",
		tandoor.WithAccessToken("tda_your_token_here"),
	)
	if err != nil {
		log.Fatal(err)
	}

	opts := &recipe.ListOptions{
		ListOptions: pagination.ListOptions{PageSize: 10, Search: "chicken"},
	}
	page, err := client.Recipes().List(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range page.Results {
		fmt.Println(r.ID, r.Name)
	}
}
```

Key API surface:

- **`tandoor.NewClient(baseURL, opts...)`** — construct a client. Options include
  `WithAccessToken`, `WithHTTPClient` (e.g., to plug in a recording transport), and
  `WithBaseURL`. The default HTTP timeout is 30s.
- **Service accessors** — `client.Recipes()`, `client.Foods()`, `client.Ingredients()`,
  `client.Steps()`, `client.Units()`, `client.UnitConversions()`, `client.ShoppingLists()`,
  `client.ShoppingEntries()`, `client.MealPlans()`, `client.CookLogs()`,
  `client.RecipeBooks()`, `client.Properties()`, `client.Spaces()`, `client.Users()`,
  `client.Storages()`, `client.Supermarkets()`, `client.RecipeImports()`,
  `client.RecipeFromSource()`, `client.ShareLinks()`, and many more — one per Tandoor API
  resource, grouped under domain subpackages (`recipe`, `food`, `shopping`, `mealplan`,
  `space`, `action`, `importexport`, ...).
- **Pagination** — list methods return `*pagination.Paginated[T]` (also aliased as
  `tandoor.Paginated[T]`): `Count`, `Next`, `Previous`, `Results`, with helpers like
  `HasNext()`. Pass `pagination.ListOptions` (page, page size, search, ordering) to filter.
- **Errors** — API failures return `*tandoor.TandoorError` with `StatusCode`, `Body`, and a
  parsed `Message`; `errors.Is(err, &tandoor.TandoorError{})` matches any API error, or
  match a specific status code via a `TandoorError{StatusCode: 404}` target.
- **`fdc` package** — a standalone client for the USDA [FoodData Central](https://fdc.nal.usda.gov/)
  API (search, get by ID, batch, paged listing):

  ```go
  fdcClient, err := fdc.NewClient(fdc.WithAPIKey("your-key"))
  food, err := fdcClient.GetFood(ctx, 534358, fdc.FormatFull, nil) // nil = all nutrients
  ```

## Development

```sh
go build ./...          # build
go test ./...           # test
go run ./cmd/tandoor-cli --help
```

`go test ./...` runs unit tests only. The integration and CLI end-to-end suites spin up a
disposable Tandoor instance and require `podman-compose` on PATH; see
`internal/tests/README.md` and `internal/cli/e2e/README.md`.

Layout:

```
.
├── client.go          # root tandoor package: Client, options, service accessors
├── error.go           # TandoorError
├── pagination/        # Paginated[T] and ListOptions
├── <domain>/          # typed services: recipe, food, shopping, mealplan, ...
├── fdc/               # USDA FoodData Central client
├── detector/          # food-name issue detectors (audit support)
├── normalize/         # food-name normalization (audit support)
├── internal/          # shared HTTP executor; integration & CLI e2e test suites
└── cmd/tandoor-cli/   # the CLI (urfave/cli v3)
```

The API surface tracks the Tandoor REST API as documented in the [Tandoor documentation](https://docs.tandoor.me/).

## License

MIT — see [LICENSE](LICENSE).
