// cmd/tandoor-cli/cmd_audit_food.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/auditfood"
	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/urfave/cli/v3"
)

// auditFoodCommand returns the `audit food <id>` subcommand group.
func auditFoodCommand() *cli.Command {
	return &cli.Command{
		Name:      "food",
		Usage:     "Inspect, suggest, or fix a single food",
		ArgsUsage: "id",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "fdc-api-key",
				Usage: "USDA FDC API key (defaults to FDC_API_KEY env var)",
			},
		},
		Commands: []*cli.Command{
			auditFoodInspect(),
			auditFoodSuggest(),
			auditFoodFix(),
		},
	}
}

// auditFoodInspect returns the `audit food inspect <id>` subcommand.
func auditFoodInspect() *cli.Command {
	return &cli.Command{
		Name:        "inspect",
		Usage:       "Deep-dive on one food: details, ingredients, issues, FDC lookup",
		Description: "Returns food details, ingredients usage, detector issues, suggested normalized name and FDC matches for manual review.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}

			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			var fdcClient *fdc.Client
			if fdcClient, err = newFdcClient(cmd); err != nil {
				fdcClient = nil // FDC lookup is optional
			}

			result, err := auditfood.Inspect(ctx, c, fdcClient, id, cmd.Int("page-size"))
			if err != nil {
				return err
			}
			return printJSON(result)
		},
	}
}

// auditFoodSuggest returns the `audit food suggest <id>` subcommand.
func auditFoodSuggest() *cli.Command {
	return &cli.Command{
		Name:        "suggest",
		Usage:       "Propose a correction (name + FDC ID) for a food",
		Description: "Shows proposed normalized name, FDC candidate and affected ingredients before applying changes.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}

			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			inspect, err := auditfood.Inspect(ctx, c, nil, id, cmd.Int("page-size"))
			if err != nil {
				return err
			}
			f := inspect.Food
			norm := normalize.Normalize(f.Name)

			type AffectedIngredient struct {
				ID         int    `json:"id"`
				RecipeID   int    `json:"recipe_id"`
				RecipeName string `json:"recipe_name"`
			}
			var affectedIngredients []AffectedIngredient
			for _, ing := range inspect.Ingredients {
				for _, ref := range ing.UsedInRecipes {
					affectedIngredients = append(affectedIngredients, AffectedIngredient{
						ID:         ing.ID,
						RecipeID:   ref.ID,
						RecipeName: ref.Name,
					})
				}
			}
			if affectedIngredients == nil {
				affectedIngredients = []AffectedIngredient{}
			}

			type Change struct {
				Field string `json:"field"`
				From  any    `json:"from"`
				To    any    `json:"to"`
			}
			type SuggestResult struct {
				FoodID              int                  `json:"food_id"`
				CurrentName         string               `json:"current_name"`
				ProposedName        string               `json:"proposed_name"`
				ProposedFDCID       *int                 `json:"proposed_fdc_id"`
				Changes             []Change             `json:"changes"`
				IssuesResolved      []string             `json:"issues_resolved"`
				PrepNote            string               `json:"prep_note,omitempty"`
				Alternatives        []string             `json:"alternatives,omitempty"`
				AffectedIngredients []AffectedIngredient `json:"affected_ingredients"`
			}

			result := SuggestResult{
				FoodID:       id,
				CurrentName:  f.Name,
				ProposedName: norm.Canonical,
				Changes: []Change{
					{
						Field: "name",
						From:  f.Name,
						To:    norm.Canonical,
					},
				},
				IssuesResolved:      inspect.Issues,
				PrepNote:            norm.PrepNote,
				Alternatives:        norm.Alternatives,
				AffectedIngredients: affectedIngredients,
			}

			// If no FDC ID is set, add a change for it (pending lookup).
			if f.FDCID == nil {
				result.Changes = append(result.Changes, Change{
					Field: "fdc_id",
					From:  nil,
					To:    "pending_lookup",
				})
			}

			return printJSON(result)
		},
	}
}

// auditFoodFix returns the `audit food fix <id>` subcommand.
func auditFoodFix() *cli.Command {
	return &cli.Command{
		Name:        "fix",
		Usage:       "Apply the correction to a food (write)",
		Description: "Applies normalized name change to food (merging into an identically named food on collision). Requires --yes flag. Use --dry-run to preview.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Print the plan without sending",
			},
			&cli.BoolFlag{
				Name:  "yes",
				Usage: "Required to apply the fix (agents should use --yes)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}

			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			dryRun := cmd.Bool("dry-run")
			yes := cmd.Bool("yes")

			if !dryRun && !yes {
				return fmt.Errorf("--yes flag is required to apply the fix (use --dry-run to preview)")
			}

			result, err := auditfood.Fix(ctx, c, &auditfood.FixOptions{
				FoodID:           id,
				MergeOnCollision: true,
			}, dryRun)
			if err != nil {
				return err
			}
				switch {
			case dryRun:
				fmt.Fprintln(cmd.ErrWriter, "[dry-run] plan for food", id)
			case result.Merged:
				fmt.Fprintf(cmd.ErrWriter, "[fix] food %d merged into %d\n", id, result.Collision.ID)
			default:
				fmt.Fprintf(cmd.ErrWriter, "[fix] food %d: %q -> %q\n", id, result.OriginalName, result.NewName)
			}
			return printJSON(result)
		},
	}
}
