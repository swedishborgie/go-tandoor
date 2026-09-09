# go-tandoor

A Go client library for the [Tandoor Recipes](https://tandoorRecipes.com/) REST API, plus
`tandoor`, a full-featured command-line client for managing recipes, shopping lists,
meal plans, and food data — built for humans *and* AI agents.

- **Library** (`github.com/swedishborgie/go-tandoor`): typed, paginated access to the Tandoor
  REST API, covering recipes, ingredients, steps, foods, units & conversions, shopping lists,
  meal plans, cook logs, recipe books, properties, spaces, imports, and more.
- **CLI** (`tandoor`): a single binary that wraps the API with sensible subcommands,
  batch operations, JSON output (`--jq` filters, `--output-file`), dry-run mode for writes,
  food-data audit tooling, and a client for the USDA FoodData Central (FDC) API. The same
  binary also runs as an MCP server via `tandoor mcp`.
- **MCP server** (`tandoor mcp`): a [Model Context Protocol](https://modelcontextprotocol.io/) server
  exposing the API as ~200 tools for AI agents, with optional parameters,
  a read-only mode, and stdio or streamable-HTTP transports. Embeddable via the `mcp` package.

## Installation

### CLI

```sh
go install github.com/swedishborgie/go-tandoor/cmd/tandoor@latest
```

The CLI binary also serves as the MCP server (`tandoor mcp`) — one install covers both.

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
tandoor --base-url https://recipes.example.com auth token --username you --password secret

# OIDC browser login (opens your browser, catches the callback on localhost:9999)
tandoor --base-url https://recipes.example.com auth oidc --oidc-backend github
```

Both print the token to stdout. The CLI also has `auth list-tokens` for enumerating existing tokens.

### 2. Make requests

```sh
export TANDOOR_TOKEN=tda_...
export TANDOOR_BASE_URL=https://recipes.example.com   # required: your Tandoor instance URL

tandoor recipes list --search "chicken"
tandoor recipes get 42
tandoor foods ensure --name "butternut squash" --name "thyme"
tandoor shopping add-recipe 42 --list-id 1
```

Output is pretty-printed JSON. List commands support `--all` to collect every page into one
array, and many support `--jq '<filter>'` (built-in [gojq](https://github.com/itchyny/gojq), no
external `jq` binary needed) and `--output-file`.

Write commands support `--dry-run` to preview the exact payload without sending it.

## CLI reference

```
tandoor [global options] <command> [command options]
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

Run `tandoor <command> --help` (or any subcommand) for flags and usage.

### Food-data audit

The `audit` group scans your food database for hygiene issues and can propose/apply
corrections, optionally cross-checking against USDA FoodData Central:

```sh
tandoor audit foods --category quantity_prefix   # scan for naming issues
tandoor audit duplicates --threshold 0.9         # find near-duplicate foods
tandoor audit food inspect 123                   # deep-dive one food (needs FDC_API_KEY)
tandoor audit food suggest 123                   # propose a correction
tandoor audit food fix 123 --yes                 # apply it (--dry-run to preview)
tandoor audit connectors                         # analyze +/-/& ingredient lines
```

## MCP server

`tandoor mcp` exposes the Tandoor API to AI agents as MCP tools: ~200 tools covering every
domain (recipes, ingredients, steps, foods, units, properties, shopping, meal plans, cook
logs, books, imports, spaces/users, FDC, and audit composites). Design notes:

- **Optional parameters with sensible defaults** (e.g. `page_size` 50), so most calls work
  with a single argument.
- **`jq` on every read tool** (built-in gojq) to project fields and keep responses small.
- **`all=true`** on paginated list tools auto-paginates and returns a flat array.
- **Write tools perform real writes**; destructive ones carry the `destructive` annotation so clients can gate them.
- **Read-only mode** (`--read-only`) registers read tools only; write tools are not exposed.
- **Tool filter** (`--tools`) allows `name` / `prefix_*` and denies with `-name`.

### Usage

One binary operates as either CLI or MCP server:

```sh
# stdio transport (default): speaks MCP over stdin/stdout; logs go to stderr
TANDOOR_BASE_URL=https://recipes.example.com TANDOOR_TOKEN=tda_... tandoor mcp

# streamable HTTP transport: POST/GET http://127.0.0.1:8090/mcp
tandoor mcp --transport http --http-addr 127.0.0.1:8090
```

Configuration (flags or env vars):

| Env var | Flag | Description |
| --- | --- | --- |
| `TANDOOR_BASE_URL` | `--base-url` | Tandoor instance base URL (**required**) |
| `TANDOOR_TOKEN` | `--token` | API access token (`tda_...`) (**required**) |
| `FDC_API_KEY` | `--fdc-api-key` | Enables the `fdc_*` tools (optional) |
| `TANDOOR_PAGE_SIZE` | `--page-size` | Default page size for list tools (default `50`) |
| `TANDOOR_MCP_HTTP_ADDR` | `--http-addr` | Listen address for the http transport (default `127.0.0.1:8090`) |
| `TANDOOR_MCP_READ_ONLY` | `--read-only` | Register read tools only (write tools are not exposed) |
| `TANDOOR_MCP_TOOLS` | `--tools` | Tool filter: `name` or `prefix_*` to allow, `-name` to deny |

Claude Desktop / Claude Code / pi accept a command-based server entry, e.g.
`claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "tandoor": {
      "command": "tandoor",
      "args": ["mcp"],
      "env": {
        "TANDOOR_BASE_URL": "https://recipes.example.com",
        "TANDOOR_TOKEN": "tda_..."
      }
    }
  }
}
```

The server is also embeddable in your own Go program:

```go
import (
    mcpgoserver "github.com/mark3labs/mcp-go/server"
    mcpsrv "github.com/swedishborgie/go-tandoor/mcp"
)

s := mcpsrv.NewServer(tandoorClient, fdcClient, mcpsrv.WithReadOnly())
mcpgoserver.ServeStdio(s) // or mcpgoserver.NewStreamableHTTPServer(s)
```

### Tool catalog

The catalog below is generated from the tool registration code — run
`go generate ./mcp` to regenerate it after changing tools, and `go test ./mcp` to verify it
matches (the test fails when the README drifts).

<!-- BEGIN GENERATED: MCP tool catalog (go generate ./mcp) -->

#### Recipes

| Tool | Description |
| --- | --- |
| `recipe_add_to_shopping` | Add a recipe's ingredients to a shopping list. With list_recipe, edits that existing entry instead; servings 0 with list_recipe deletes it. |
| `recipe_ai_properties` | Trigger server-side AI to generate keywords, servings, and times for a recipe. Requires an AI provider configured on the instance. |
| `recipe_batch_update` | Update multiple recipes at once: set keywords and/or working/waiting time. Returns the updated recipes. |
| `recipe_create` | Create a recipe from a raw Tandoor recipe payload. Returns the created recipe. |
| `recipe_delete` | Delete a recipe. |
| `recipe_delete_external` | Remove the external file reference from a recipe (keeps the recipe). |
| `recipe_get` | Get a single recipe by ID, including ingredients, steps, and keywords. |
| `recipe_list` | List recipes. Supports text query, keyword/space/book/user filters, ordering, and pagination. Use all=true for the full set; jq projects fields to keep output small. |
| `recipe_overview` | List a lightweight overview of recipes (id, name, image only — no ingredients/steps). Supports text query; returns the full set as a single array. Use jq to project fields. |
| `recipe_patch` | Partially update a recipe. Only fields present in data change. |
| `recipe_related` | List recipes related to the given recipe. Use all=true for the full set; jq projects fields to keep output small. |
| `recipe_update` | Replace a recipe with the given data (full replacement; use recipe_patch for partial changes). Returns the updated recipe. |
| `recipe_upload_image` | Set or replace a recipe's image from a URL or base64 data. |

#### Ingredients

| Tool | Description |
| --- | --- |
| `ingredient_create` | Create a standalone ingredient. Note: the API only shows ingredients attached to a visible recipe's steps, so prefer creating ingredients inside a recipe (recipe_create/update); use ingredient_patch/delete on recipe ingredients. Returns the created ingredient. |
| `ingredient_delete` | Delete an ingredient. Fails if it is still in use. |
| `ingredient_get` | Get a single ingredient by ID (includes its food, unit, and amount). |
| `ingredient_list` | List ingredients. Filter by recipe, food, or space. Use all=true for the full set; jq projects fields to keep output small. |
| `ingredient_patch` | Partially update an ingredient. Only provided fields change. |
| `ingredient_update` | Update an ingredient (full replacement). Returns the updated ingredient. |

#### Steps

| Tool | Description |
| --- | --- |
| `step_create` | Create a standalone step. Note: the API only shows steps attached to a visible recipe, so use step_update/patch/delete on steps created inside a recipe (via recipe_create/update). Returns the created step. |
| `step_delete` | Delete a step (removes it from its recipe). |
| `step_get` | Get a single step by ID. |
| `step_list` | List steps. Filter by recipe. Use all=true for the full set; jq projects fields to keep output small. |
| `step_patch` | Partially update a step. Only provided fields change. |
| `step_update` | Update a step (full replacement). Returns the updated step. |

#### Foods

| Tool | Description |
| --- | --- |
| `food_ai_properties` | Trigger server-side AI to generate properties for a food. Requires an AI provider configured on the instance. |
| `food_batch_update` | Batch update foods: add, remove, or replace substitutes for a set of foods at once. |
| `food_create` | Create a food. Returns the created food. |
| `food_delete` | Delete a food. Fails if the food is still in use. |
| `food_ensure` | Ensure foods exist by exact name, creating missing ones (composite of food_list + food_create + FDC candidate lookup). With force_create=false it only reports which names exist, match, or would be created. |
| `food_fdc_import` | Pull USDA FDC data into a food that already has an fdc_id set (populates properties and conversions server-side). |
| `food_get` | Get a single food by ID (includes category, unit, and properties). |
| `food_list` | List foods. Use query for fuzzy name search, name_exact for a case-insensitive exact name, or names for a batch exact lookup returning a {name: food\|null} map. Filters: category_id, unit_id. Use all=true for the full set; jq projects fields to keep output small. |
| `food_merge` | Merge one food into another: usages re-point to the target, then the source is deleted. |
| `food_move` | Move a food under a new parent in the food tree. |
| `food_patch` | Partially update a food. Only provided fields change. |
| `food_update` | Update a food (full replacement). Returns the updated food. |

#### Keywords

| Tool | Description |
| --- | --- |
| `keyword_create` | Create a keyword. Returns the created keyword. |
| `keyword_delete` | Delete a keyword. Recipes keeping it are re-pointed to its parent. |
| `keyword_get` | Get a single keyword by ID. |
| `keyword_list` | List keywords. Filter by recipe book. Use all=true for the full set; jq projects fields to keep output small. |
| `keyword_merge` | Merge one keyword into another: recipes move to the target, then the source is deleted. |
| `keyword_move` | Move a keyword under a new parent (re-parent in the keyword tree). |
| `keyword_patch` | Partially update a keyword. Only provided fields change. |
| `keyword_update` | Update a keyword (full replacement). Returns the updated keyword. |

#### Units & Conversions

| Tool | Description |
| --- | --- |
| `unit_conversion_create` | Create a unit conversion. Returns the created conversion. |
| `unit_conversion_delete` | Delete a unit conversion. |
| `unit_conversion_get` | Get a single unit conversion by ID. |
| `unit_conversion_list` | List unit conversions (how much of a unit a food weighs/volumes). Filter by food. Use all=true for the full set; jq projects fields to keep output small. |
| `unit_conversion_patch` | Partially update a unit conversion. Note: the API still requires base/converted unit and amounts. |
| `unit_conversion_update` | Update a unit conversion. All conversion fields are required by the API. Returns the updated conversion. |
| `unit_create` | Create a unit of measurement. Returns the created unit. |
| `unit_delete` | Delete a unit. Fails if the unit is still in use. |
| `unit_get` | Get a single unit by ID. |
| `unit_list` | List units of measurement. Use all=true for the full set; jq projects fields to keep output small. Unit IDs are instance-specific — enumerate before referencing them in writes. |
| `unit_merge` | Merge one unit into another: usages re-point to the target, then the source is deleted. |
| `unit_patch` | Partially update a unit. Only provided fields change. |
| `unit_update` | Update a unit (full replacement). Returns the updated unit. |

#### Properties

| Tool | Description |
| --- | --- |
| `property_attach` | Attach a nutrient property to a food per 100 g (composite: resolves the food, sets the per-100 unit, creates or updates the property). Idempotent — re-running updates the existing property. |
| `property_create` | Create a property value. Returns the created property. Property type IDs are instance-specific — use property_type_list first. |
| `property_delete` | Delete a property value. |
| `property_get` | Get a single property value by ID. |
| `property_list` | List property values attached to foods (e.g. calories per 100 g). Filter by food. Use all=true for the full set; jq projects fields to keep output small. |
| `property_patch` | Partially update a property value. Only provided fields change. |
| `property_type_create` | Create a property type (e.g. calories, protein). Returns the created type. |
| `property_type_delete` | Delete a property type. Fails if property values still use it. |
| `property_type_get` | Get a single property type by ID. |
| `property_type_list` | List property types (e.g. Calories, Protein). Types define which properties foods can have; their IDs are instance-specific — enumerate before referencing them in writes. |
| `property_type_patch` | Partially update a property type. Only provided fields change. |
| `property_type_update` | Update a property type (full replacement). Returns the updated type. |
| `property_update` | Update a property value (full replacement). Returns the updated property. |

#### Recipe Books

| Tool | Description |
| --- | --- |
| `book_create` | Create a recipe book. Returns the created book. |
| `book_delete` | Delete a recipe book (its recipe links are removed). |
| `book_entry_create` | Add a recipe to a recipe book. Returns the created entry. |
| `book_entry_delete` | Remove a recipe from a recipe book. |
| `book_entry_list` | List recipe book entries (book↔recipe links). Filter by book. Use all=true for the full set; jq projects fields to keep output small. |
| `book_get` | Get a single recipe book by ID (includes its recipes). |
| `book_list` | List recipe books. Use all=true for the full set; jq projects fields to keep output small. |
| `book_update` | Update a recipe book (full replacement). Returns the updated book. |

#### Shopping

| Tool | Description |
| --- | --- |
| `shopping_entry_bulk_update` | Bulk update shopping list entries: check/uncheck and move between lists in one call. |
| `shopping_entry_create` | Create a shopping list entry. Returns the created entry. |
| `shopping_entry_delete` | Delete a shopping list entry. |
| `shopping_entry_get` | Get a single shopping entry by ID. |
| `shopping_entry_list` | List shopping entries (items to buy). Use all=true for the full set; jq projects fields to keep output small. |
| `shopping_entry_patch` | Partially update a shopping list entry. Only provided fields change. |
| `shopping_entry_update` | Update a shopping list entry (full replacement). Returns the updated entry. |
| `shopping_list_add_recipe` | Add a recipe's ingredients to a shopping list (creates a shopping list recipe and entries). Amounts scale with servings vs the recipe's servings. |
| `shopping_list_create` | Create a shopping list. Returns the created list. |
| `shopping_list_delete` | Delete a shopping list (its entries are removed). |
| `shopping_list_get` | Get a single shopping list by ID (includes its entries and recipes). |
| `shopping_list_list` | List shopping lists. Use all=true for the full set; jq projects fields to keep output small. |
| `shopping_list_update` | Update a shopping list (full replacement). Returns the updated list. |
| `shopping_recipe_create_entries` | Create shopping entries from a recipe's ingredients and attach them to shopping lists (composite: derives scaled entries, creates a shopping list recipe when needed). Pass shopping_list_recipe_id to reuse an existing link, or recipe_id to create a new one. |
| `shopping_recipe_list` | List shopping list recipes (recipes contributing ingredients to shopping lists). Use all=true for the full set; jq projects fields to keep output small. |

#### Meal Plans

| Tool | Description |
| --- | --- |
| `meal_plan_auto_plan` | Auto-generate meal plans for a date range; Tandoor picks the recipes matching the keywords server-side. |
| `meal_plan_create` | Create a meal plan entry. Returns the created entry. Meal type IDs are instance-specific — use meal_type_list first. |
| `meal_plan_delete` | Delete a meal plan entry. |
| `meal_plan_get` | Get a single meal plan entry by ID. |
| `meal_plan_ical` | Export meal plans as an iCal/ICS calendar (plain text). Filter by space, user, and/or date range. |
| `meal_plan_list` | List meal plans. Filter by space, user, and/or a date range (date_from/date_to, YYYY-MM-DD). Use all=true for the full set; jq projects fields to keep output small. |
| `meal_plan_update` | Update a meal plan entry (full replacement). Returns the updated entry. |
| `meal_type_create` | Create a meal type (e.g. breakfast). Returns the created meal type. |
| `meal_type_delete` | Delete a meal type. Fails if meal plans still use it. |
| `meal_type_get` | Get a single meal type by ID. |
| `meal_type_list` | List meal types (Breakfast, Lunch, Dinner, ...). Their IDs are instance-specific — enumerate before referencing them in writes. |
| `meal_type_patch` | Partially update a meal type. Only provided fields change. |
| `meal_type_update` | Update a meal type (full replacement). Returns the updated meal type. |

#### Cook Logs

| Tool | Description |
| --- | --- |
| `cook_log_create` | Log a cooked meal (cook log entry). Returns the created cook log. |
| `cook_log_get` | Get a single cook log entry by ID. |
| `cook_log_list` | List cook logs (recorded times a recipe was cooked). Filter by recipe. Use all=true for the full set; jq projects fields to keep output small. |

#### Inventory

| Tool | Description |
| --- | --- |
| `inventory_entry_create` | Create an inventory entry (item on hand). Returns the created entry. |
| `inventory_entry_delete` | Delete an inventory entry. |
| `inventory_entry_get` | Get a single inventory entry by ID. |
| `inventory_entry_list` | List inventory entries (food quantities on hand). Filter by food, location, or barcode; include_empty controls zero-quantity rows. Use all=true for the full set; jq projects fields to keep output small. |
| `inventory_entry_patch` | Partially update an inventory entry. Only provided fields change. |
| `inventory_entry_update` | Update an inventory entry (full replacement). Returns the updated entry. |
| `inventory_location_create` | Create an inventory location (e.g. fridge, pantry). Returns the created location. |
| `inventory_location_delete` | Delete an inventory location. Fails if entries still reference it. |
| `inventory_location_get` | Get a single inventory location by ID. |
| `inventory_location_list` | List inventory locations (fridge, pantry, ...). Use all=true for the full set; jq projects fields to keep output small. |
| `inventory_location_patch` | Partially update an inventory location. Only provided fields change. |
| `inventory_location_update` | Update an inventory location (full replacement). Returns the updated location. |
| `inventory_log_list` | List inventory logs (stock changes). Filter by entry or food. Use all=true for the full set; jq projects fields to keep output small. |

#### Storages

| Tool | Description |
| --- | --- |
| `storage_create` | Create a storage backend (DB, NEXTCLOUD, or LOCAL). Returns the created storage. |
| `storage_delete` | Delete a storage backend. Fails if files still reference it. |
| `storage_get` | Get a single storage by ID. |
| `storage_list` | List storages (legacy storage locations; prefer inventory locations for new data). Use all=true for the full set; jq projects fields to keep output small. |
| `storage_patch` | Partially update a storage backend. Only provided fields change. |
| `storage_update` | Update a storage backend (full replacement). Returns the updated storage. |

#### Supermarkets

| Tool | Description |
| --- | --- |
| `supermarket_category_get` | Get a single supermarket category by ID. |
| `supermarket_category_list` | List supermarket categories. Use all=true for the full set; jq projects fields to keep output small. |
| `supermarket_get` | Get a single supermarket by ID. |
| `supermarket_list` | List supermarkets. Use all=true for the full set; jq projects fields to keep output small. |

#### Imports & Sharing

| Tool | Description |
| --- | --- |
| `bookmarklet_import_list` | List bookmarklet imports (recipes captured via the browser bookmarklet). Use all=true for the full set; jq projects fields to keep output small. |
| `export_log_list` | List export logs (history of recipe exports). Use all=true for the full set; jq projects fields to keep output small. |
| `import_delete` | Delete a staged recipe import without executing it. |
| `import_get` | Get a single recipe import by ID (the staged payload). |
| `import_import` | Execute a single staged recipe import (creates the recipe). |
| `import_import_all` | Execute all pending staged recipe imports. |
| `import_list` | List recipe imports (staged imports waiting to be executed). Use all=true for the full set; jq projects fields to keep output small. |
| `import_log_list` | List import logs (history of executed imports). Use all=true for the full set; jq projects fields to keep output small. |
| `open_data_import` | Import USDA open data (foods, units, conversions) for a version and datatypes. Check open_data_metadata for valid values first. |
| `open_data_metadata` | List available USDA open data versions and datatypes (call before open_data_import). |
| `recipe_from_source_create` | Create a recipe from a source URL or raw HTML/JSON payload (server-side scrape). The response includes detected duplicates when the URL was imported before. |
| `share_link_create` | Create a public share link for a recipe. Returns the share URL. |

#### Syncs

| Tool | Description |
| --- | --- |
| `sync_create` | Create a sync configuration (external recipe sync source). Returns the created sync configuration. |
| `sync_delete` | Delete a sync configuration. |
| `sync_get` | Get a single sync configuration by ID. |
| `sync_list` | List sync configurations (external recipe sync sources). Use all=true for the full set; jq projects fields to keep output small. |
| `sync_log_list` | List sync logs (history of sync runs). Use all=true for the full set; jq projects fields to keep output small. |
| `sync_patch` | Partially update a sync configuration. Only provided fields change. |
| `sync_query_synced_folder` | Trigger a sync run for a sync configuration's folder and return the resulting sync log. |
| `sync_update` | Replace a sync configuration with the given data (full replacement). Returns the updated sync configuration. |

#### Spaces, Users & Groups

| Tool | Description |
| --- | --- |
| `group_get` | Get a single permission group by ID. |
| `group_list` | List permission groups. Returns the full set as a single array. |
| `household_create` | Create a household in the current space. |
| `household_delete` | Delete a household. |
| `household_get` | Get a single household by ID. |
| `household_list` | List households. Supports text query, ordering, and pagination. Use all=true for the full set; jq projects fields to keep output small. |
| `household_patch` | Partially update a household's name (the only writable field). |
| `household_update` | Update a household's name (the only writable field). Returns the updated household. |
| `invite_link_create` | Create an invite link to join the space with a given group. Returns the invite URL. |
| `invite_link_delete` | Delete an invite link. |
| `invite_link_get` | Get a single invite link by ID (includes the invite URL). |
| `invite_link_list` | List space invite links (unused by default; set used=true to include used links). |
| `space_current` | Get the space bound to the current access token (the default space for this token). |
| `space_get` | Get a single space by ID. |
| `space_list` | List spaces in the Tandoor instance. Supports text query, ordering, and pagination. Use all=true for the full set; jq projects fields to keep output small. |
| `user_get` | Get a single user by ID. |
| `user_list` | List users. Optionally restrict to members of the given space IDs (space_ids). Returns the full set as a single array. |
| `user_patch` | Update a user's first/last name. Only these fields are writable via the API. |

#### Auth

| Tool | Description |
| --- | --- |
| `auth_create_token` | Create a new API access token for the current user. The token value is only returned in full immediately after creation. |
| `auth_delete_token` | Delete an API access token. |
| `auth_get_token` | Get a single API access token by ID. |
| `auth_list_tokens` | List API access tokens for the current user (with scopes and expiry). |

#### Misc & Admin

| Tool | Description |
| --- | --- |
| `ai_log_get` | Get a single AI log entry by ID. |
| `ai_log_list` | List AI logs (AI feature usage history). Use all=true for the full set; jq projects fields to keep output small. |
| `ai_provider_get` | Get a single AI provider configuration by ID. |
| `ai_provider_list` | List AI provider configurations. Use all=true for the full set; jq projects fields to keep output small. |
| `automation_get` | Get a single automation by ID. |
| `automation_list` | List automations (scheduled tasks). Use all=true for the full set; jq projects fields to keep output small. |
| `connector_config_get` | Get a single connector configuration by ID. |
| `connector_config_list` | List connector configurations (external service integrations). Use all=true for the full set; jq projects fields to keep output small. |
| `custom_filter_get` | Get a single custom filter by ID. |
| `custom_filter_list` | List custom filters (saved recipe search filters). Use all=true for the full set; jq projects fields to keep output small. |
| `localization_list` | List available localizations (language/locale settings); returns the full set as a single array. |
| `search_field_list` | List custom search fields (returns the full set as a single array). Use jq to project fields. |
| `search_preference_list` | List search preferences (per-user default search settings; returns the full set as a single array). Use jq to project fields. |
| `search_preference_patch` | Update a user's search preferences (data passes through verbatim; fields: search, lookup, trigram_threshold, unaccent/icontains/istartswith/trigram/fulltext as lists of {name, field} objects). |
| `server_settings_get` | Get the current server settings (read-only view of instance-wide settings). |
| `user_file_get` | Get a single user file's metadata by ID. |
| `user_file_list` | List user files (uploaded attachments). Use all=true for the full set; jq projects fields to keep output small. |
| `view_log_get` | Get a single view log entry by ID. |
| `view_log_list` | List view logs (recipe view history). Use all=true for the full set; jq projects fields to keep output small. |

#### Audit Composites

| Tool | Description |
| --- | --- |
| `food_audit_fix` | Rename a food to its normalized canonical name (Title Case, prep words stripped). On a name collision it merges into the existing food unless merge=false. Use food_audit_fix_preview to see the plan first. |
| `food_audit_fix_preview` | Preview the food_audit_fix plan without writing: normalized name, alternatives, and the collision/merge decision. Read-only. |
| `food_audit_inspect` | Deep-dive on one food: details, ingredient usage, naming issues, suggested canonical name, and FDC candidates when the food has no FDC ID. Read-only. |
| `food_find_duplicates` | Scan all foods and group names that are likely duplicates (normalized Jaccard word similarity). Read-only. |

#### FoodData Central (FDC)

| Tool | Description |
| --- | --- |
| `fdc_get_food` | Get a single FDC food with full nutrient detail by FDC ID (from fdc_search). |
| `fdc_search` | Search the USDA FoodData Central database by name. Returns FDC IDs and abridged nutrient data; pair with fdc_get_food for full detail. |

#### Server

| Tool | Description |
| --- | --- |
| `server_info` | Server status: Tandoor base URL, tool count, and read-only/FDC configuration. Call first to understand the environment. |

<!-- END GENERATED: MCP tool catalog -->

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
go run ./cmd/tandoor --help
```

### Versioning

`tandoor --version` reports the version injected at build time. Releases are tag-driven:
pushing a `v*` tag runs GoReleaser (`.github/workflows/release.yml`), which builds
linux/windows × amd64/arm64 binaries and creates the GitHub release. Locally, build with
the git-derived version (tag + commit, `-dirty` when uncommitted):

```sh
go build -ldflags "-X main.version=$(git describe --tags --dirty --always --long)" -o tandoor ./cmd/tandoor
```

`go test ./...` runs unit tests only. The integration and CLI end-to-end suites spin up a
disposable Tandoor instance and require `podman-compose` or the `docker compose` v2 plugin on PATH; see
`internal/tests/README.md`, `internal/cli/e2e/README.md`, and `internal/mcp/e2e/README.md`. They run on GitHub Actions (`integration.yml`) with the `docker` fallback.

Layout:

```
.
├── client.go          # root tandoor package: Client, options, service accessors
├── error.go           # TandoorError
├── pagination/        # Paginated[T] and ListOptions
├── <domain>/          # typed services: recipe, food, shopping, mealplan, ...
├── fdc/               # USDA FoodData Central client
├── mcp/               # embeddable MCP server library + tool catalog
├── detector/          # food-name issue detectors (audit support)
├── normalize/         # food-name normalization (audit support)
├── internal/          # shared HTTP executor; integration & CLI e2e test suites
└── cmd/tandoor/   # the CLI (urfave/cli v3), including the "mcp" MCP-server subcommand
```

The API surface tracks the Tandoor REST API as documented in the [Tandoor documentation](https://docs.tandoor.me/).

## License

MIT — see [LICENSE](LICENSE).
