// cmd/tandoor-cli/cmd_foods.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/swedishborgie/go-tandoor/unit"
	"github.com/urfave/cli/v3"
)

// GetFoodsCommand returns the top-level `foods` command group.
func GetFoodsCommand() *cli.Command {
	return &cli.Command{
		Name:  "foods",
		Usage: "Food operations",
		Commands: []*cli.Command{
			foodsListCommand(),
			foodsGetCommand(),
			foodsCreateCommand(),
			foodsUpdateCommand(),
			foodsPatchCommand(),
			foodsDeleteCommand(),
			foodsMergeCommand(),
			foodsEnsureCommand(),
			foodsAttachFDCPropertiesCommand(),
			foodsAutoConversionsCommand(),
		},
	}
}

func foodsListCommand() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "List foods (GET /api/food/)",
		Description: "List foods with optional search, exact name match, or batch name lookup. Supports --all pagination, --jq filtering and --output-file.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
			&cli.StringFlag{
				Name:  "name-exact",
				Usage: "Exact food name match",
			},
			&cli.StringSliceFlag{
				Name:  "name",
				Usage: "Food names to lookup (batch, exact match)",
			},
			&cli.IntFlag{
				Name:  "category-id",
				Usage: "Filter by category ID",
			},
			&cli.IntFlag{
				Name:  "unit-id",
				Usage: "Filter by unit ID",
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Collect all pages into a single array",
			},
			&cli.StringFlag{
				Name:  "jq",
				Usage: "jq filter for output",
			},
			&cli.StringFlag{
				Name:  "output-file",
				Usage: "Write output to file",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			search := cmd.String("search")
			nameExact := cmd.String("name-exact")
			nameSlice := cmd.StringSlice("name")
			// Batch name lookup
			if len(nameSlice) > 0 {
				results := make(map[string]*food.Food)
				for _, name := range nameSlice {
					o := &food.ListOptions{ListOptions: pagination.ListOptions{PageSize: 100, Search: name}}
					page, err := c.Foods().List(ctx, o)
					if err != nil {
						printError(err)
						return err
					}
					var found *food.Food
					for _, f := range page.Results {
						if strings.EqualFold(f.Name, name) {
							found = &f
							break
						}
					}
					results[name] = found
				}
				return outputWithJQ(ctx, results, cmd.String("jq"), cmd.String("output-file"))
			}
			pageSize := cmd.Int("page-size")

			opts := &food.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize}}
			if search != "" {
				opts.Search = search
			}
			if cmd.IsSet("category-id") {
				opts.CategoryID = cmd.Int("category-id")
			}
			if cmd.IsSet("unit-id") {
				opts.UnitID = cmd.Int("unit-id")
			}
			if nameExact != "" {
				if opts.Extra == nil {
					opts.Extra = map[string]string{}
				}
				// The food endpoint supports `query` for name contains lookup.
				// We use that to narrow results, then enforce exact matching client-side.
				opts.Extra["query"] = nameExact

				matches := make([]food.Food, 0)
				pageNum := 1
				for {
					opts.Page = pageNum
					page, err := c.Foods().List(ctx, opts)
					if err != nil {
						printError(err)
						return err
					}
					for _, f := range page.Results {
						if strings.EqualFold(f.Name, nameExact) {
							matches = append(matches, f)
						}
					}
					if !page.HasNext() {
						break
					}
					pageNum++
				}
				return outputWithJQ(ctx, matches, cmd.String("jq"), cmd.String("output-file"))
			}

			page, err := c.Foods().List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[food.Food], error) {
					opts.Page = pageNum
					return c.Foods().List(ctx, opts)
				})
				if err != nil {
					printError(err)
					return err
				}
				return outputWithJQ(ctx, all, cmd.String("jq"), cmd.String("output-file"))
			}
			return outputWithJQ(ctx, page, cmd.String("jq"), cmd.String("output-file"))
		},
	}
}

func foodsGetCommand() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "Get a food by ID (GET /api/food/<id>/)",
		Description: "Fetch a single food by ID or batch fetch multiple IDs via --id flag. Returns full food object with properties and FDC ID.",
		ArgsUsage:   "id",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "id",
				Usage: "Food IDs to get (batch)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			idsSlice := cmd.StringSlice("id")
			if len(idsSlice) > 0 {
				results := make(map[string]*food.Food)
				for _, s := range idsSlice {
					id, err := parseIDArg(s)
					if err != nil {
						printError(err)
						return err
					}
					f, err := c.Foods().Get(ctx, id)
					if err != nil {
						printError(err)
						return err
					}
					results[fmt.Sprintf("%d", id)] = f
				}
				return printJSON(results)
			}
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			f, err := c.Foods().Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(f)
		},
	}
}

func foodsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a new food (POST /api/food/)",
		ArgsUsage: "name",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview payload without sending",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			name := cmd.Args().First()
			if name == "" {
				return fmt.Errorf("food name is required")
			}
			payload := &food.Food{Name: name}
			if cmd.Bool("dry-run") {
				fmt.Fprintf(cmd.ErrWriter, "[dry-run] POST api/food/ payload:\n")
				return printJSON(payload)
			}
			f, err := c.Foods().Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] food %q id=%d\n", f.Name, f.ID)
			return printJSON(f)
		},
	}
}

func foodFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "plural-name",
			Usage: "Plural display name",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Food description",
		},
		&cli.StringFlag{
			Name:  "shopping",
			Usage: "Shopping behavior value",
		},
		&cli.BoolFlag{
			Name:  "ignore-shopping",
			Usage: "Ignore food in shopping list generation",
		},
		&cli.StringFlag{
			Name:  "open-data-slug",
			Usage: "Open data slug",
		},
		&cli.IntFlag{
			Name:  "parent",
			Usage: "Parent food ID",
		},
		&cli.IntFlag{
			Name:  "fdc-id",
			Usage: "USDA FDC ID",
		},
		&cli.BoolFlag{
			Name:  "clear-fdc-id",
			Usage: "Clear USDA FDC ID",
		},
		&cli.BoolFlag{
			Name:  "substitute-onhand",
			Usage: "Enable substitute-on-hand behavior",
		},
	}
}

func applyFoodFlags(cmd *cli.Command, f *food.Food) {
	if cmd.IsSet("name") {
		f.Name = cmd.String("name")
	}
	if cmd.IsSet("plural-name") {
		f.PluralName = cmd.String("plural-name")
	}
	if cmd.IsSet("description") {
		f.Description = cmd.String("description")
	}
	if cmd.IsSet("shopping") {
		f.Shopping = cmd.String("shopping")
	}
	if cmd.IsSet("ignore-shopping") {
		f.IgnoreShopping = cmd.Bool("ignore-shopping")
	}
	if cmd.IsSet("open-data-slug") {
		f.OpenDataSlug = cmd.String("open-data-slug")
	}
	if cmd.IsSet("parent") {
		f.Parent = cmd.Int("parent")
	}
	if cmd.IsSet("fdc-id") {
		fdcID := cmd.Int("fdc-id")
		f.FDCID = &fdcID
	}
	if cmd.IsSet("clear-fdc-id") && cmd.Bool("clear-fdc-id") {
		f.FDCID = nil
	}
	if cmd.IsSet("substitute-onhand") {
		f.SubstituteOnHand = cmd.Bool("substitute-onhand")
	}
}

func foodsUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a food (PUT /api/food/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Food name",
			},
		}, foodFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			payload, err := c.Foods().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get food %d: %w", id, err)
			}
			applyFoodFlags(cmd, payload)

			updated, err := c.Foods().Update(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] food %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func foodsPatchCommand() *cli.Command {
	return &cli.Command{
		Name:        "patch",
		Usage:       "Patch a food (PATCH /api/food/<id>/)",
		Description: "Partially update a food. Supports --name and --fdc-id. Sends minimal payload to avoid overwriting fields.",
		ArgsUsage:   "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Food name",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview PATCH payload without sending",
			},
		}, foodFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			// Build a minimal PATCH payload from the flags that were set,
			// avoiding GET-then-PATCH which can lose fields under concurrent
			// requests (e.g. fdc_id nullified when properties attach runs).
			payload, err := buildFoodPatchPayload(cmd)
			if err != nil {
				printError(err)
				return err
			}

			if cmd.Bool("dry-run") {
				fmt.Fprintf(cmd.ErrWriter, "[dry-run] PATCH api/food/%d/ payload:\n", id)
				return printJSON(payload)
			}

			var updated map[string]any
			if err := c.DoJSON(ctx, "PATCH", "api/food/"+fmt.Sprintf("%d/", id), payload, &updated); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[patch] food %d updated\n", id)
			return printJSON(updated)
		},
	}
}

// buildFoodPatchPayload builds a minimal map[string]any from the flags set
// on the command, without needing a prior GET of the full food.
func buildFoodPatchPayload(cmd *cli.Command) (map[string]any, error) {
	payload := make(map[string]any)
	if cmd.IsSet("name") {
		payload["name"] = cmd.String("name")
	}
	if cmd.IsSet("plural-name") {
		payload["plural_name"] = cmd.String("plural-name")
	}
	if cmd.IsSet("description") {
		payload["description"] = cmd.String("description")
	}
	if cmd.IsSet("shopping") {
		payload["shopping"] = cmd.String("shopping")
	}
	if cmd.IsSet("ignore-shopping") {
		payload["ignore_shopping"] = cmd.Bool("ignore-shopping")
	}
	if cmd.IsSet("open-data-slug") {
		payload["open_data_slug"] = cmd.String("open-data-slug")
	}
	if cmd.IsSet("parent") {
		payload["parent"] = cmd.Int("parent")
	}
	if cmd.IsSet("fdc-id") {
		payload["fdc_id"] = cmd.Int("fdc-id")
	}
	if cmd.IsSet("clear-fdc-id") && cmd.Bool("clear-fdc-id") {
		payload["fdc_id"] = nil
	}
	if cmd.IsSet("substitute-onhand") {
		payload["substitute_onhand"] = cmd.Bool("substitute-onhand")
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("no fields to patch — provide at least one flag")
	}
	return payload, nil
}

func foodsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a food (DELETE /api/food/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			if err := c.Foods().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] food %d\n", id)
			return nil
		},
	}
}

func foodsMergeCommand() *cli.Command {
	return &cli.Command{
		Name:        "merge",
		Usage:       "Merge one food into another (PUT /api/food/<source>/merge/<target>/)",
		Description: "Reassign all ingredients from source food to target food and delete source. Use when a canonical food already exists to avoid name collisions.",
		ArgsUsage:   "source_id target_id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)

			args := cmd.Args().Slice()
			if len(args) < 2 {
				return fmt.Errorf("source_id and target_id are required")
			}

			sourceID, err := parseIDArg(args[0])
			if err != nil {
				return fmt.Errorf("invalid source_id: %w", err)
			}
			targetID, err := parseIDArg(args[1])
			if err != nil {
				return fmt.Errorf("invalid target_id: %w", err)
			}

			// Fetch both foods to show what's happening.
			source, err := c.Foods().Get(ctx, sourceID)
			if err != nil {
				return fmt.Errorf("get source food %d: %w", sourceID, err)
			}
			target, err := c.Foods().Get(ctx, targetID)
			if err != nil {
				return fmt.Errorf("get target food %d: %w", targetID, err)
			}

			fmt.Fprintf(cmd.ErrWriter, "Merge: %q (%d) -> %q (%d)\n",
				source.Name, sourceID, target.Name, targetID)

			_, err = c.Foods().Merge(ctx, sourceID, targetID)
			if err != nil {
				return fmt.Errorf("merge food %d into %d: %w", sourceID, targetID, err)
			}

			// Tandoor's merge endpoint returns id=0; fetch the actual target food.
			updatedTarget, err := c.Foods().Get(ctx, targetID)
			if err != nil {
				return fmt.Errorf("get target food %d after merge: %w", targetID, err)
			}

			fmt.Fprintf(cmd.ErrWriter, "Merged. Target food is now: %q (%d)\n", updatedTarget.Name, updatedTarget.ID)
			return printJSON(updatedTarget)
		},
	}
}

func foodsEnsureCommand() *cli.Command {
	return &cli.Command{
		Name:  "ensure",
		Usage: "Ensure foods exist by name, return IDs and near matches (batch via --name)",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "name",
				Usage:    "Food name to ensure (repeatable)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview actions without creating foods",
			},
			&cli.BoolFlag{
				Name:  "force-create",
				Usage: "Create food even if ambiguous (skip near-match check)",
			},
			&cli.IntFlag{
				Name:  "merge-into",
				Usage: "If ambiguous, merge into this food ID",
			},
			&cli.StringFlag{
				Name:  "output-file",
				Usage: "Write JSON output to file",
			},
		},
		Action: foodsEnsureAction,
	}
}

type ensureResult struct {
	Status        string      `json:"status"`
	Food          *food.Food  `json:"food,omitempty"`
	NearMatches   []nearMatch `json:"near_matches,omitempty"`
	FdcCandidates []any       `json:"fdc_candidates,omitempty"`
	ActionsTaken  []string    `json:"actions_taken,omitempty"`
}

type nearMatch struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Similarity float64 `json:"similarity"`
}

func foodsEnsureAction(ctx context.Context, cmd *cli.Command) error {
	c := ctx.Value(ctxKeyClient).(*tandoor.Client)
	names := cmd.StringSlice("name")
	dryRun := cmd.Bool("dry-run")
	forceCreate := cmd.Bool("force-create")

	fdcClient, err := newFdcClient(cmd)
	if err != nil {
		return err
	}

	results := make(map[string]ensureResult, len(names))

	for _, name := range names {
		res := ensureResult{
			Status:       "not_found",
			ActionsTaken: []string{},
		}
		// Search for foods with query matching name.
		opts := &food.ListOptions{
			ListOptions: pagination.ListOptions{
				PageSize: 100,
				Search:   name,
			},
		}
		page, err := c.Foods().List(ctx, opts)
		if err != nil {
			return fmt.Errorf("foods list for %q: %w", name, err)
		}

		var exact *food.Food
		var candidates []food.Food
		for _, f := range page.Results {
			if strings.EqualFold(f.Name, name) {
				exact = &f
				break
			}
			candidates = append(candidates, f)
		}

		if exact != nil {
			res.Status = "found"
			res.Food = exact
			results[name] = res
			continue
		}

		// Compute near matches via Jaccard similarity.
		near := []nearMatch{}
		for _, f := range candidates {
			sim := jaccardSimilarity(name, f.Name)
			if sim >= 0.6 {
				near = append(near, nearMatch{
					ID:         f.ID,
					Name:       f.Name,
					Similarity: sim,
				})
			}
		}
		for i := 0; i < len(near)-1; i++ {
			for j := i + 1; j < len(near); j++ {
				if near[j].Similarity > near[i].Similarity {
					near[i], near[j] = near[j], near[i]
				}
			}
		}

		if len(near) > 0 {
			res.Status = "ambiguous"
			res.NearMatches = near
		}

		// FDC candidates for agent selection
		type fdcCandidate struct {
			FDCID       int     `json:"fdc_id"`
			Description string  `json:"description"`
			DataType    string  `json:"data_type"`
			Score       float64 `json:"score,omitempty"`
		}
		limit := 5
		pageSize := limit
		if pageSize > 200 {
			pageSize = 200
		}
		nilInt := (*int)(nil)
		fdcResp, fdcErr := fdcClient.SearchFoods(ctx, name, nil, &pageSize, nilInt, "", "", "")
		if fdcErr == nil {
			candidatesFDC := []fdcCandidate{}
			for _, f := range fdcResp.Foods {
				if len(candidatesFDC) >= limit {
					break
				}
				c := fdcCandidate{
					FDCID:       f.FDCID,
					Description: f.Description,
					DataType:    f.DataType,
				}
				if f.Score != nil {
					c.Score = *f.Score
				}
				candidatesFDC = append(candidatesFDC, c)
			}
			anySlice := make([]any, len(candidatesFDC))
			for i, v := range candidatesFDC {
				anySlice[i] = v
			}
			res.FdcCandidates = anySlice
		}

		if forceCreate {
			if !dryRun {
				newFood := &food.Food{Name: name}
				created, err := c.Foods().Create(ctx, newFood)
				if err != nil {
					return fmt.Errorf("create food %q: %w", name, err)
				}
				res.Status = "created"
				res.Food = created
				res.ActionsTaken = append(res.ActionsTaken, "food_created")
			} else {
				res.Status = "would_create"
				res.ActionsTaken = append(res.ActionsTaken, "food_create_dry_run")
			}
			results[name] = res
			continue
		}

		res.NearMatches = near
		results[name] = res
	}

	// Idempotency summary
	summary := struct {
		Total       int `json:"total"`
		Found       int `json:"found"`
		Created     int `json:"created"`
		WouldCreate int `json:"would_create"`
		Ambiguous   int `json:"ambiguous"`
		NotFound    int `json:"not_found"`
	}{}
	summary.Total = len(names)
	for _, r := range results {
		switch r.Status {
		case "found":
			summary.Found++
		case "created":
			summary.Created++
		case "would_create":
			summary.WouldCreate++
		case "ambiguous":
			summary.Ambiguous++
		default:
			summary.NotFound++
		}
	}
	output := map[string]any{
		"summary": summary,
		"results": results,
	}
	if outFile := cmd.String("output-file"); outFile != "" {
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal output: %w", err)
		}
		if err := os.WriteFile(outFile, data, 0644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		fmt.Fprintf(cmd.ErrWriter, "[output] wrote results to %s\n", outFile)
		fmt.Fprintf(cmd.ErrWriter, "[summary] total=%d found=%d created=%d would_create=%d ambiguous=%d not_found=%d\n", summary.Total, summary.Found, summary.Created, summary.WouldCreate, summary.Ambiguous, summary.NotFound)
	}
	return printJSON(output)
}

func foodsAttachFDCPropertiesCommand() *cli.Command {
	return &cli.Command{
		Name:  "attach-fdc-properties",
		Usage: "Attach properties to a food from USDA FDC data based on property type fdc_id",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "food-id",
				Usage:    "Food ID",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "fdc-id",
				Usage: "Override FDC ID (defaults to food.FDCID)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview actions without creating properties",
			},
		},
		Action: foodsAttachFDCPropertiesAction,
	}
}

func foodsAutoConversionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "auto-conversions",
		Usage: "Create common unit conversions for a food based on unit heuristics",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "food-id",
				Usage:    "Food ID",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview conversions without creating",
			},
		},
		Action: foodsAutoConversionsAction,
	}
}

func foodsAttachFDCPropertiesAction(ctx context.Context, cmd *cli.Command) error {
	c := ctx.Value(ctxKeyClient).(*tandoor.Client)
	foodID := cmd.Int("food-id")
	fdcOverride := cmd.Int("fdc-id")
	dryRun := cmd.Bool("dry-run")

	foodObj, err := c.Foods().Get(ctx, foodID)
	if err != nil {
		return fmt.Errorf("get food %d: %w", foodID, err)
	}

	fdcIDPtr := foodObj.FDCID
	if cmd.IsSet("fdc-id") {
		fdcID := fdcOverride
		fdcIDPtr = &fdcID
	}
	if fdcIDPtr == nil {
		return fmt.Errorf("food %d has no FDC ID and --fdc-id was not provided", foodID)
	}
	fdcID := *fdcIDPtr

	fdcClient, err := newFdcClient(cmd)
	if err != nil {
		return err
	}
	fdcFood, err := fdcClient.GetFood(ctx, fdcID, "", nil)
	if err != nil {
		return fmt.Errorf("fdc get %d: %w", fdcID, err)
	}

	// Load property types with fdc_id set
	ptSvc := property.NewTypeService(c)
	ptPage, err := ptSvc.List(ctx, &property.TypeListOptions{ListOptions: pagination.ListOptions{PageSize: 500}})
	if err != nil {
		return fmt.Errorf("list property types: %w", err)
	}

	attached := []map[string]any{}
	for _, pt := range ptPage.Results {
		if pt.FDCID == nil {
			continue
		}
		fdcNutrientID := *pt.FDCID
		// Find matching nutrient in FDC food
		var amount *float64
		for _, fn := range fdcFood.FoodNutrients {
			if fn.Nutrient == nil {
				continue
			}
			// Match by nutrient number (string) or ID
			matched := false
			if fn.Nutrient.Number != "" {
				if fmt.Sprintf("%d", fdcNutrientID) == fn.Nutrient.Number {
					matched = true
				}
			}
			if fn.Nutrient.ID != 0 && int(fn.Nutrient.ID) == fdcNutrientID {
				matched = true
			}
			if matched && fn.Amount != nil {
				amount = fn.Amount
				break
			}
		}
		if amount == nil {
			continue
		}
		entry := map[string]any{
			"property_type_id":   pt.ID,
			"property_type_name": pt.Name,
			"fdc_id":             fdcNutrientID,
			"amount":             *amount,
		}
		if dryRun {
			attached = append(attached, entry)
			continue
		}
		// Attach property via existing attach logic: PATCH food with properties array
		// Fetch food to get current properties
		foodMap := map[string]any{}
		if err := c.DoJSON(ctx, "GET", "api/food/"+fmt.Sprintf("%d/", foodID), nil, &foodMap); err != nil {
			return fmt.Errorf("get food %d for patch: %w", foodID, err)
		}
		props, _ := foodMap["properties"].([]any)
		// Check if property already exists
		exists := false
		for _, p := range props {
			pm, ok := p.(map[string]any)
			if !ok {
				continue
			}
			ptMap, ok := pm["property_type"].(map[string]any)
			if !ok {
				continue
			}
			if int(ptMap["id"].(float64)) == pt.ID {
				exists = true
				// Update amount
				pm["property_amount"] = *amount
				break
			}
		}
		if !exists {
			newProp := map[string]any{
				"property_type": map[string]any{
					"id":   pt.ID,
					"name": pt.Name,
				},
				"property_amount": *amount,
			}
			props = append(props, newProp)
		}
		patch := map[string]any{
			"properties":             props,
			"properties_food_amount": 100,
			"properties_food_unit":   17,
		}
		if err := c.DoJSON(ctx, "PATCH", "api/food/"+fmt.Sprintf("%d/", foodID), patch, &map[string]any{}); err != nil {
			return fmt.Errorf("patch food %d properties: %w", foodID, err)
		}
		attached = append(attached, entry)
	}

	if dryRun {
		fmt.Fprintf(cmd.ErrWriter, "[dry-run] would attach %d properties to food %d\n", len(attached), foodID)
		return printJSON(attached)
	}
	fmt.Fprintf(cmd.ErrWriter, "[attach] attached %d properties to food %d\n", len(attached), foodID)
	return printJSON(attached)
}

func foodsAutoConversionsAction(ctx context.Context, cmd *cli.Command) error {
	c := ctx.Value(ctxKeyClient).(*tandoor.Client)
	foodID := cmd.Int("food-id")
	dryRun := cmd.Bool("dry-run")

	foodObj, err := c.Foods().Get(ctx, foodID)
	if err != nil {
		return fmt.Errorf("get food %d: %w", foodID, err)
	}
	// Use properties_food_unit as heuristic base unit if set, otherwise default to gram (17)
	var unitID int
	if foodObj.PropertiesFoodUnit != nil {
		// PropertiesFoodUnit may be int or map
		switch v := foodObj.PropertiesFoodUnit.(type) {
		case float64:
			unitID = int(v)
		case int:
			unitID = v
		case map[string]any:
			if idf, ok := v["id"].(float64); ok {
				unitID = int(idf)
			}
		}
	}
	if unitID == 0 {
		unitID = 17 // default gram
	}

	// Load units to determine category
	unitsPage, err := c.Units().List(ctx, &unit.ListOptions{ListOptions: pagination.ListOptions{PageSize: 500}})
	if err != nil {
		return fmt.Errorf("list units: %w", err)
	}
	var unitName string
	for _, u := range unitsPage.Results {
		if u.ID == unitID {
			unitName = u.Name
			break
		}
	}

	// Heuristic mapping
	type convPair struct {
		BaseUnitID      int
		BaseAmount      float64
		ConvertedUnitID int
		ConvertedAmount float64
	}
	var pairs []convPair
	lower := strings.ToLower(unitName)
	if strings.Contains(lower, "gram") || strings.Contains(lower, "g ") {
		// weight -> oz, lb, kg
		pairs = []convPair{
			{BaseUnitID: unitID, BaseAmount: 28.3495, ConvertedUnitID: 18, ConvertedAmount: 1}, // g -> oz
			{BaseUnitID: unitID, BaseAmount: 453.592, ConvertedUnitID: 19, ConvertedAmount: 1}, // g -> lb
			{BaseUnitID: unitID, BaseAmount: 1000, ConvertedUnitID: 20, ConvertedAmount: 1},    // g -> kg
		}
	} else if strings.Contains(lower, "millilitre") || strings.Contains(lower, "ml") {
		pairs = []convPair{
			{BaseUnitID: unitID, BaseAmount: 240, ConvertedUnitID: 21, ConvertedAmount: 1}, // ml -> cup
			{BaseUnitID: unitID, BaseAmount: 15, ConvertedUnitID: 22, ConvertedAmount: 1},  // ml -> tbsp
			{BaseUnitID: unitID, BaseAmount: 5, ConvertedUnitID: 23, ConvertedAmount: 1},   // ml -> tsp
		}
	}

	if len(pairs) == 0 {
		return fmt.Errorf("no heuristic conversions for unit %q", unitName)
	}

	created := []map[string]any{}
	for _, p := range pairs {
		// Check if conversion already exists
		listOpts := &unit.ConversionListOptions{ListOptions: &pagination.ListOptions{PageSize: 500}}
		foodIDPtr := foodID
		listOpts.FoodID = &foodIDPtr
		convPage, err := c.UnitConversions().List(ctx, listOpts)
		if err != nil {
			return fmt.Errorf("list conversions: %w", err)
		}
		exists := false
		for _, cv := range convPage.Results {
			if cv.BaseUnit != nil && cv.ConvertedUnit != nil && cv.Food != nil && cv.Food.ID == foodID {
				if cv.BaseUnit.ID == p.BaseUnitID && cv.ConvertedUnit.ID == p.ConvertedUnitID {
					exists = true
					break
				}
			}
		}
		if exists {
			continue
		}
		conv := &unit.Conversion{
			BaseUnit:        &unit.Ref{ID: p.BaseUnitID},
			BaseAmount:      p.BaseAmount,
			ConvertedUnit:   &unit.Ref{ID: p.ConvertedUnitID},
			ConvertedAmount: p.ConvertedAmount,
			Food:            &unit.FoodRef{ID: foodID},
		}
		if dryRun {
			created = append(created, map[string]any{
				"base_unit_id":      p.BaseUnitID,
				"base_amount":       p.BaseAmount,
				"converted_unit_id": p.ConvertedUnitID,
				"converted_amount":  p.ConvertedAmount,
			})
			continue
		}
		createdConv, err := c.UnitConversions().Create(ctx, conv)
		if err != nil {
			return fmt.Errorf("create conversion: %w", err)
		}
		created = append(created, map[string]any{
			"id":                createdConv.ID,
			"base_unit_id":      p.BaseUnitID,
			"base_amount":       p.BaseAmount,
			"converted_unit_id": p.ConvertedUnitID,
			"converted_amount":  p.ConvertedAmount,
		})
	}

	if dryRun {
		fmt.Fprintf(cmd.ErrWriter, "[dry-run] would create %d conversions for food %d\n", len(created), foodID)
	} else {
		fmt.Fprintf(cmd.ErrWriter, "[auto-conversions] created %d conversions for food %d\n", len(created), foodID)
	}
	return printJSON(created)
}

func jaccardSimilarity(a, b string) float64 {
	setA := wordSet(strings.ToLower(a))
	setB := wordSet(strings.ToLower(b))
	if len(setA) == 0 && len(setB) == 0 {
		return 1.0
	}
	intersection := 0
	for w := range setA {
		if _, ok := setB[w]; ok {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
