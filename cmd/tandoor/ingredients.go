// cmd/tandoor/cmd_ingredients.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/urfave/cli/v3"
)

// GetIngredientsCommand returns the top-level `ingredients` command group.
func GetIngredientsCommand() *cli.Command {
	return &cli.Command{
		Name:  "ingredients",
		Usage: "Ingredient operations",
		Commands: []*cli.Command{
			ingredientsListCommand(),
			ingredientsGetCommand(),
			ingredientsCreateCommand(),
			ingredientsUpdateCommand(),
			ingredientsPatchCommand(),
			ingredientsDeleteCommand(),
		},
	}
}

func ingredientsListCommand() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "List ingredients (GET /api/ingredient/)",
		Description: "List ingredients, optionally filtered by recipe-id. Use --all for full pagination.",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "recipe-id",
				Usage: "Filter ingredients for a recipe ID",
			},
			&cli.IntFlag{
				Name:  "food-id",
				Usage: "Filter ingredients by food ID",
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Collect all pages into a single array",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			pageSize := cmd.Int("page-size")

			if cmd.IsSet("recipe-id") {
				recipeID := cmd.Int("recipe-id")
				r, err := c.Recipes().Get(ctx, recipeID)
				if err != nil {
					return fmt.Errorf("get recipe %d: %w", recipeID, err)
				}
				filtered := make([]ingredient.Ingredient, 0)
				for _, st := range r.Steps {
					for _, ing := range st.Ingredients {
						amount := ing.Amount
						filtered = append(filtered, ingredient.Ingredient{
							ID:           ing.ID,
							Amount:       &amount,
							Note:         ing.Note,
							Order:        ing.Order,
							IsHeader:     ing.IsHeader,
							NoAmount:     ing.NoAmount,
							OriginalText: ing.OriginalText,
							Checked:      ing.Checked,
						})
						if ing.Food != nil {
							filtered[len(filtered)-1].Food = &ingredient.Food{
								ID:   ing.Food.ID,
								Name: ing.Food.Name,
							}
						}
						if ing.Unit != nil {
							filtered[len(filtered)-1].Unit = &ingredient.Unit{
								ID:   ing.Unit.ID,
								Name: ing.Unit.Name,
							}
						}
					}
				}
				return printJSON(filtered)
			}

			opts := &ingredient.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize}}
			if cmd.IsSet("food-id") {
				opts.FoodID = cmd.Int("food-id")
			}
			page, err := c.Ingredients().List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[ingredient.Ingredient], error) {
					opts.Page = pageNum
					return c.Ingredients().List(ctx, opts)
				})
				if err != nil {
					printError(err)
					return err
				}
				return printJSON(all)
			}
			return outputWithJQ(ctx, page, cmd.String("jq"), cmd.String("output-file"))
		},
	}
}

func ingredientsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get an ingredient by ID (GET /api/ingredient/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			i, err := c.Ingredients().Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(i)
		},
	}
}

func ingredientUpdateFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "food",
			Usage: "Food ID",
		},
		&cli.IntFlag{
			Name:  "unit",
			Usage: "Unit ID",
		},
		&cli.FloatFlag{
			Name:  "amount",
			Usage: "Amount",
		},
		&cli.StringFlag{
			Name:  "note",
			Usage: "Note text",
		},
		&cli.IntFlag{
			Name:  "order",
			Usage: "Step order",
		},
		&cli.BoolFlag{
			Name:  "is-header",
			Usage: "Set ingredient as a section header",
		},
		&cli.BoolFlag{
			Name:  "no-amount",
			Usage: "Hide amount for this ingredient",
		},
		&cli.StringFlag{
			Name:  "original-text",
			Usage: "Original free-form ingredient text",
		},
		&cli.BoolFlag{
			Name:  "checked",
			Usage: "Mark ingredient as checked",
		},
	}
}

func applyIngredientFlags(ctx context.Context, c *tandoor.Client, cmd *cli.Command, payload *ingredient.Ingredient) error {
	if cmd.IsSet("food") {
		foodID := cmd.Int("food")
		newFood, err := c.Foods().Get(ctx, foodID)
		if err != nil {
			return fmt.Errorf("get food %d: %w", foodID, err)
		}
		payload.Food = &ingredient.Food{ID: foodID, Name: newFood.Name}
	}
	if cmd.IsSet("unit") {
		unitID := cmd.Int("unit")
		u, err := c.Units().Get(ctx, unitID)
		if err != nil {
			return fmt.Errorf("get unit %d: %w", unitID, err)
		}
		payload.Unit = &ingredient.Unit{ID: unitID, Name: u.Name}
	}
	if cmd.IsSet("amount") {
		amount := cmd.Float("amount")
		payload.Amount = &amount
	}
	if cmd.IsSet("note") {
		payload.Note = cmd.String("note")
	}
	if cmd.IsSet("order") {
		payload.Order = cmd.Int("order")
	}
	if cmd.IsSet("is-header") {
		payload.IsHeader = cmd.Bool("is-header")
	}
	if cmd.IsSet("no-amount") {
		payload.NoAmount = cmd.Bool("no-amount")
	}
	if cmd.IsSet("original-text") {
		payload.OriginalText = cmd.String("original-text")
	}
	if cmd.IsSet("checked") {
		payload.Checked = cmd.Bool("checked")
	}
	return nil
}

func ingredientsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create an ingredient (POST /api/ingredient/)",
		ArgsUsage: "",
		Flags: append(ingredientUpdateFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Preview payload without sending",
		}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)

			payload := &ingredient.Ingredient{}
			if err := applyIngredientFlags(ctx, c, cmd, payload); err != nil {
				return err
			}

			if cmd.Bool("dry-run") {
				fmt.Fprintf(cmd.ErrWriter, "[dry-run] POST api/ingredient/ payload:\n")
				return printJSON(payload)
			}

			created, err := c.Ingredients().Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] ingredient id=%d\n", created.ID)
			return printJSON(created)
		},
	}
}

func ingredientsUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update an ingredient (PUT /api/ingredient/<id>/)",
		ArgsUsage: "id",
		Flags:     ingredientUpdateFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			// Fetch current ingredient
			current, err := c.Ingredients().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get ingredient %d: %w", id, err)
			}

			payload := *current
			if err := applyIngredientFlags(ctx, c, cmd, &payload); err != nil {
				return err
			}

			updated, err := c.Ingredients().Update(ctx, &payload)
			if err != nil {
				return fmt.Errorf("update ingredient %d: %w", id, err)
			}

			if payload.Food != nil && current.Food != nil {
				fmt.Fprintf(cmd.ErrWriter, "[update] ingredient %d food: %q (%d) -> %q (%d)\n",
					id, current.Food.Name, current.Food.ID, updated.Food.Name, updated.Food.ID)
			} else {
				fmt.Fprintf(cmd.ErrWriter, "[update] ingredient %d updated\n", id)
			}
			return printJSON(updated)
		},
	}
}

func ingredientsPatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Patch an ingredient (PATCH /api/ingredient/<id>/)",
		ArgsUsage: "id",
		Flags: append(ingredientUpdateFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Preview PATCH payload without sending",
		}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			current, err := c.Ingredients().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get ingredient %d: %w", id, err)
			}

			payload := *current
			if err := applyIngredientFlags(ctx, c, cmd, &payload); err != nil {
				return err
			}

			if cmd.Bool("dry-run") {
				fmt.Fprintf(cmd.ErrWriter, "[dry-run] PATCH api/ingredient/%d/ payload:\n", id)
				return printJSON(payload)
			}

			updated, err := c.Ingredients().Patch(ctx, &payload)
			if err != nil {
				return fmt.Errorf("patch ingredient %d: %w", id, err)
			}
			fmt.Fprintf(cmd.ErrWriter, "[patch] ingredient %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func ingredientsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete an ingredient (DELETE /api/ingredient/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			if err := c.Ingredients().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] ingredient %d\n", id)
			return nil
		},
	}
}
