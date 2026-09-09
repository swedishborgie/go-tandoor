// cmd/tandoor/cmd_foods.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/auditfood"
	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
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
			&cli.StringFlag{
				Name:  "output-file",
				Usage: "Write JSON output to file",
			},
		},
		Action: foodsEnsureAction,
	}
}

func foodsEnsureAction(ctx context.Context, cmd *cli.Command) error {
	c := ctx.Value(ctxKeyClient).(*tandoor.Client)
	names := cmd.StringSlice("name")
	dryRun := cmd.Bool("dry-run")
	forceCreate := cmd.Bool("force-create")

	var fdcClient *fdc.Client
	if fc, err := newFdcClient(cmd); err == nil {
		fdcClient = fc
	} else {
		// FDC candidates are optional — ensure still works without a key.
		fmt.Fprintf(cmd.ErrWriter, "[ensure] no FDC API key, skipping FDC candidates\n")
	}

	report, err := auditfood.Ensure(ctx, c, &auditfood.EnsureOptions{
		Names:       names,
		FDC:         fdcClient,
		ForceCreate: forceCreate,
	}, dryRun)
	if err != nil {
		return err
	}
	output := map[string]any{
		"summary": report.Summary,
		"results": report.Results,
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
		fmt.Fprintf(cmd.ErrWriter, "[summary] total=%d found=%d created=%d would_create=%d ambiguous=%d not_found=%d\n",
			report.Summary.Total, report.Summary.Found, report.Summary.Created, report.Summary.WouldCreate, report.Summary.Ambiguous, report.Summary.NotFound)
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
	dryRun := cmd.Bool("dry-run")

	fdcClient, err := newFdcClient(cmd)
	if err != nil {
		return err
	}

	opts := &auditfood.FdcAttachOptions{FoodID: foodID}
	if cmd.IsSet("fdc-id") {
		id := cmd.Int("fdc-id")
		opts.FDCID = &id
	}

	result, err := auditfood.AttachFDCProperties(ctx, c, fdcClient, opts, dryRun)
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Fprintf(cmd.ErrWriter, "[dry-run] would attach %d properties to food %d\n", len(result.Attached), foodID)
	} else {
		fmt.Fprintf(cmd.ErrWriter, "[attach] attached %d properties to food %d\n", len(result.Attached), foodID)
	}
	return printJSON(result)
}

func foodsAutoConversionsAction(ctx context.Context, cmd *cli.Command) error {
	c := ctx.Value(ctxKeyClient).(*tandoor.Client)
	foodID := cmd.Int("food-id")
	dryRun := cmd.Bool("dry-run")

	result, err := auditfood.AutoConversions(ctx, c, foodID, dryRun)
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Fprintf(cmd.ErrWriter, "[dry-run] would create %d conversions for food %d\n", len(result.Created), foodID)
	} else {
		fmt.Fprintf(cmd.ErrWriter, "[auto-conversions] created %d conversions for food %d\n", len(result.Created), foodID)
	}
	return printJSON(result)
}
