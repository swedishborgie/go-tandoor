// cmd/tandoor-cli/cmd_recipes.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/urfave/cli/v3"
)

// GetRecipesCommand returns the top-level `recipes` command group.
func GetRecipesCommand() *cli.Command {
	return &cli.Command{
		Name:  "recipes",
		Usage: "Recipe operations",
		Commands: []*cli.Command{
			recipesListCommand(),
			recipesGetCommand(),
			recipesCreateCommand(),
			recipesUpdateCommand(),
			recipesPatchCommand(),
			recipesDeleteCommand(),
			recipesOverviewCommand(),
			recipesBatchUpdateCommand(),
			recipesSetImageCommand(),
			recipesRelatedCommand(),
		},
	}
}

func recipesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List recipes (GET /api/recipe/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
			&cli.BoolFlag{
				Name:  "flat",
				Usage: "Use flat endpoint (GET /api/recipe/flat/)",
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Collect all pages into a single array",
			},
			&cli.StringFlag{
				Name:  "keyword-ids",
				Usage: "Comma-separated keyword IDs to filter",
			},
			&cli.IntFlag{
				Name:  "space-id",
				Usage: "Filter by space ID",
			},
			&cli.IntFlag{
				Name:  "recipe-book-id",
				Usage: "Filter by recipe book ID",
			},
			&cli.IntFlag{
				Name:  "user-id",
				Usage: "Filter by user ID",
			},
			&cli.BoolFlag{
				Name:  "is-favorite",
				Usage: "Filter by favorite status",
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
			pageSize := cmd.Int("page-size")

			opts := &recipe.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize}}
			if search != "" {
				opts.Search = search
			}
			if cmd.IsSet("keyword-ids") {
				ids, err := parseIntCSV(cmd.String("keyword-ids"))
				if err != nil {
					return fmt.Errorf("parse --keyword-ids: %w", err)
				}
				opts.KeywordIDs = ids
			}
			if cmd.IsSet("space-id") {
				opts.SpaceID = cmd.Int("space-id")
			}
			if cmd.IsSet("recipe-book-id") {
				opts.RecipeBookID = cmd.Int("recipe-book-id")
			}
			if cmd.IsSet("user-id") {
				opts.UserID = cmd.Int("user-id")
			}
			if cmd.IsSet("is-favorite") {
				v := cmd.Bool("is-favorite")
				opts.IsFavorite = &v
			}

			if cmd.Bool("flat") {
				flat, err := c.Recipes().Flat(ctx, opts)
				if err != nil {
					printError(err)
					return err
				}
				return outputWithJQ(ctx, flat, cmd.String("jq"), cmd.String("output-file"))
			}

			page, err := c.Recipes().List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[recipe.Recipe], error) {
					opts.Page = pageNum
					return c.Recipes().List(ctx, opts)
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

func recipesGetCommand() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "Get a recipe by ID (GET /api/recipe/<id>/)",
		Description: "Fetch full recipe with steps, ingredients and food_properties for verification.",
		ArgsUsage:   "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			r, err := c.Recipes().Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(r)
		},
	}
}

func recipeUpsertFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "Recipe name",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Recipe description",
		},
		&cli.StringFlag{
			Name:  "keywords",
			Usage: "Comma-separated keyword IDs",
		},
		&cli.StringSliceFlag{
			Name:  "step",
			Usage: "Recipe step instruction (repeatable)",
		},
		&cli.IntFlag{
			Name:  "working-time",
			Usage: "Working time in minutes",
		},
		&cli.IntFlag{
			Name:  "waiting-time",
			Usage: "Waiting time in minutes",
		},
		&cli.IntFlag{
			Name:  "servings",
			Usage: "Number of servings",
		},
		&cli.StringFlag{
			Name:  "source-url",
			Usage: "Source URL",
		},
		&cli.BoolFlag{
			Name:  "internal",
			Usage: "Mark recipe as internal",
		},
		&cli.BoolFlag{
			Name:  "private",
			Usage: "Mark recipe as private",
		},
		&cli.BoolFlag{
			Name:  "show-ingredient-overview",
			Usage: "Show ingredient overview",
		},
		&cli.StringFlag{
			Name:  "file",
			Usage: "Load recipe payload JSON from file",
		},
	}
}

func loadRecipeFromFile(path string) (*recipe.Recipe, error) {
	if path == "" {
		return nil, nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read recipe file %q: %w", path, err)
	}
	var payload recipe.Recipe
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse recipe JSON %q: %w", path, err)
	}
	return &payload, nil
}

func applyRecipeFlags(cmd *cli.Command, payload *recipe.Recipe) error {
	if cmd.IsSet("name") {
		payload.Name = cmd.String("name")
	}
	if cmd.IsSet("description") {
		payload.Description = cmd.String("description")
	}
	if cmd.IsSet("keywords") {
		ids, err := parseIntCSV(cmd.String("keywords"))
		if err != nil {
			return fmt.Errorf("parse --keywords: %w", err)
		}
		keywords := make([]recipe.Keyword, 0, len(ids))
		for _, id := range ids {
			keywords = append(keywords, recipe.Keyword{ID: id})
		}
		payload.Keywords = keywords
	}
	if cmd.IsSet("step") {
		stepValues := cmd.StringSlice("step")
		steps := make([]recipe.Step, 0, len(stepValues))
		for i, instruction := range stepValues {
			steps = append(steps, recipe.Step{
				Order:       i + 1,
				Instruction: instruction,
			})
		}
		payload.Steps = steps
	}
	if cmd.IsSet("working-time") {
		payload.WorkingTime = cmd.Int("working-time")
	}
	if cmd.IsSet("waiting-time") {
		payload.WaitingTime = cmd.Int("waiting-time")
	}
	if cmd.IsSet("servings") {
		payload.Servings = cmd.Int("servings")
	}
	if cmd.IsSet("source-url") {
		payload.SourceURL = cmd.String("source-url")
	}
	if cmd.IsSet("internal") {
		payload.Internal = cmd.Bool("internal")
	}
	if cmd.IsSet("private") {
		payload.Private = cmd.Bool("private")
	}
	if cmd.IsSet("show-ingredient-overview") {
		payload.ShowIngredientOverview = cmd.Bool("show-ingredient-overview")
	}
	return nil
}

func recipesCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a recipe (POST /api/recipe/)",
		ArgsUsage: "",
		Flags: append(recipeUpsertFlags(),
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview payload without sending",
			},
			&cli.StringFlag{
				Name:  "output-file",
				Usage: "Write preview JSON to file",
			},
			&cli.StringFlag{
				Name:  "ingredients-file",
				Usage: "Load steps/ingredients from JSON file",
			},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			payload := &recipe.Recipe{}
			if cmd.IsSet("file") {
				fromFile, err := loadRecipeFromFile(cmd.String("file"))
				if err != nil {
					printError(err)
					return err
				}
				payload = fromFile
			}
			if err := applyRecipeFlags(cmd, payload); err != nil {
				return err
			}
			if payload.Steps == nil {
				payload.Steps = []recipe.Step{}
			}
			if cmd.IsSet("ingredients-file") {
				ingPath := cmd.String("ingredients-file")
				ingBytes, err := os.ReadFile(ingPath)
				if err != nil {
					return fmt.Errorf("read ingredients file %q: %w", ingPath, err)
				}
				var steps []recipe.Step
				if err := json.Unmarshal(ingBytes, &steps); err != nil {
					return fmt.Errorf("parse ingredients file %q: %w", ingPath, err)
				}
				payload.Steps = steps
			}
			if payload.Name == "" {
				return fmt.Errorf("recipe name is required (--name or --file)")
			}

			if cmd.Bool("dry-run") {
				fmt.Fprintf(cmd.ErrWriter, "[dry-run] POST api/recipe/ payload:\n")
				if outFile := cmd.String("output-file"); outFile != "" {
					data, err := json.MarshalIndent(payload, "", "  ")
					if err != nil {
						return fmt.Errorf("marshal dry-run payload: %w", err)
					}
					if err := os.WriteFile(outFile, data, 0644); err != nil {
						return fmt.Errorf("write output file: %w", err)
					}
					fmt.Fprintf(cmd.ErrWriter, "[output] wrote dry-run payload to %s\n", outFile)
				}
				return printJSON(payload)
			}

			created, err := c.Recipes().Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] recipe %q id=%d\n", created.Name, created.ID)
			return printJSON(created)
		},
	}
}

func recipesUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a recipe (PUT /api/recipe/<id>/)",
		ArgsUsage: "id",
		Flags:     recipeUpsertFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := c.Recipes().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get recipe %d: %w", id, err)
			}
			if cmd.IsSet("file") {
				fromFile, err := loadRecipeFromFile(cmd.String("file"))
				if err != nil {
					printError(err)
					return err
				}
				payload = fromFile
			}
			payload.ID = id
			if err := applyRecipeFlags(cmd, payload); err != nil {
				return err
			}

			updated, err := c.Recipes().Update(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] recipe %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func recipesPatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Patch a recipe (PATCH /api/recipe/<id>/)",
		ArgsUsage: "id",
		Flags:     recipeUpsertFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := c.Recipes().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get recipe %d: %w", id, err)
			}
			if cmd.IsSet("file") {
				fromFile, err := loadRecipeFromFile(cmd.String("file"))
				if err != nil {
					printError(err)
					return err
				}
				payload = fromFile
			}
			payload.ID = id
			if err := applyRecipeFlags(cmd, payload); err != nil {
				return err
			}

			updated, err := c.Recipes().Patch(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[patch] recipe %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func recipesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a recipe (DELETE /api/recipe/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			if err := c.Recipes().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] recipe %d\n", id)
			return nil
		},
	}
}

func recipesOverviewCommand() *cli.Command {
	return &cli.Command{
		Name:  "overview",
		Usage: "List recipe overview cards (GET /api/recipe-overview/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			params := url.Values{}
			params.Set("page_size", strconv.Itoa(cmd.Int("page-size")))
			if cmd.String("search") != "" {
				params.Set("search", cmd.String("search"))
			}
			path := "api/recipe-overview/"
			if qs := params.Encode(); qs != "" {
				path += "?" + qs
			}

			var page pagination.Paginated[recipe.Overview]
			if err := c.DoJSON(ctx, "GET", path, nil, &page); err != nil {
				return err
			}
			return printJSON(&page)
		},
	}
}

func recipesBatchUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:  "batch-update",
		Usage: "Batch-update recipes (PUT /api/recipe/batch_update/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "recipes",
				Usage: "Comma-separated recipe IDs (required)",
			},
			&cli.StringFlag{
				Name:  "keywords-add",
				Usage: "Comma-separated keyword IDs to add",
			},
			&cli.StringFlag{
				Name:  "keywords-remove",
				Usage: "Comma-separated keyword IDs to remove",
			},
			&cli.StringFlag{
				Name:  "keywords-set",
				Usage: "Comma-separated keyword IDs to set exactly",
			},
			&cli.BoolFlag{
				Name:  "keywords-remove-all",
				Usage: "Remove all keywords",
			},
			&cli.IntFlag{
				Name:  "working-time",
				Usage: "Set working time in minutes",
			},
			&cli.IntFlag{
				Name:  "waiting-time",
				Usage: "Set waiting time in minutes",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			recipeIDs, err := parseIntCSV(cmd.String("recipes"))
			if err != nil {
				return fmt.Errorf("parse --recipes: %w", err)
			}
			if len(recipeIDs) == 0 {
				return fmt.Errorf("--recipes is required")
			}
			keywordsAdd, err := parseIntCSV(cmd.String("keywords-add"))
			if err != nil {
				return fmt.Errorf("parse --keywords-add: %w", err)
			}
			keywordsRemove, err := parseIntCSV(cmd.String("keywords-remove"))
			if err != nil {
				return fmt.Errorf("parse --keywords-remove: %w", err)
			}
			keywordsSet, err := parseIntCSV(cmd.String("keywords-set"))
			if err != nil {
				return fmt.Errorf("parse --keywords-set: %w", err)
			}

			payload := &recipe.BatchUpdate{
				Recipes:           recipeIDs,
				KeywordsAdd:       keywordsAdd,
				KeywordsRemove:    keywordsRemove,
				KeywordsSet:       keywordsSet,
				KeywordsRemoveAll: cmd.Bool("keywords-remove-all"),
			}
			if cmd.IsSet("working-time") {
				workingTime := cmd.Int("working-time")
				payload.WorkingTime = &workingTime
			}
			if cmd.IsSet("waiting-time") {
				waitingTime := cmd.Int("waiting-time")
				payload.WaitingTime = &waitingTime
			}

			updated, err := c.Recipes().BatchUpdate(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[batch-update] recipes updated=%d\n", len(recipeIDs))
			return printJSON(updated)
		},
	}
}

func recipesSetImageCommand() *cli.Command {
	return &cli.Command{
		Name:      "set-image",
		Usage:     "Set recipe image from URL or file (PUT /api/recipe/<id>/image/)",
		ArgsUsage: "id",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url",
				Usage: "Image URL",
			},
			&cli.StringFlag{
				Name:  "file",
				Usage: "Image file value",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			imageURL := cmd.String("url")
			imageFile := cmd.String("file")
			if imageURL == "" && imageFile == "" {
				return fmt.Errorf("either --url or --file is required")
			}

			updated, err := c.Recipes().UploadImage(ctx, id, &recipe.Image{
				Image:    imageFile,
				ImageURL: imageURL,
			})
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[set-image] recipe %d image updated\n", id)
			return printJSON(updated)
		},
	}
}

func recipesRelatedCommand() *cli.Command {
	return &cli.Command{
		Name:      "related",
		Usage:     "Get related objects for a recipe (GET /api/recipe/<id>/related/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			page, err := c.Recipes().Related(ctx, id, nil)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(page)
		},
	}
}
