// cmd/tandoor-cli/cmd_audit_food.go

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/detector"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/swedishborgie/go-tandoor/pagination"
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
			f, err := c.Foods().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get food: %w", err)
			}

			// Fetch ingredients using this food (filtered by food ID).
			pageSize := cmd.Int("page-size")
			opts := &ingredient.ListOptions{ListOptions: pagination.ListOptions{
				PageSize: pageSize,
				Extra:    map[string]string{"food": fmt.Sprintf("%d", id)}},
			}
			ingredientPage, err := c.Ingredients().List(ctx, opts)
			if err != nil {
				return fmt.Errorf("list ingredients: %w", err)
			}

			foodIngredients := make([]*ingredient.Ingredient, len(ingredientPage.Results))
			for i := range ingredientPage.Results {
				foodIngredients[i] = &ingredientPage.Results[i]
			}

			// Run detectors.
			reg := newAuditRegistry(ctx, c, cmd)
			detFood := detector.Food{
				Name:            f.Name,
				FDCID:           f.FDCID,
				PropertyTypeIDs: propertyTypeIDs(f),
			}
			issues := reg.CheckAll(detFood)

			issueList := make([]string, len(issues))
			for i, iss := range issues {
				issueList[i] = fmt.Sprintf("[%s] %s (%s)", iss.Category, iss.Message, iss.Severity)
			}

			// Generate suggested name.
			norm := normalize.Normalize(f.Name)

			// FDC lookup: if fdc_id is null, search FDC with the canonical name.
			type FDCMatch struct {
				FDCID       int     `json:"fdc_id"`
				Description string  `json:"description"`
				DataType    string  `json:"data_type"`
				Score       float64 `json:"score,omitempty"`
			}
			var fdcMatches []FDCMatch
			if f.FDCID == nil {
				fdcClient, err := newFdcClient(cmd)
				if err == nil {
					limit := 5
					resp, err := fdcClient.SearchFoods(ctx, norm.Canonical, nil, &limit, nil, "", "", "")
					if err == nil {
						for _, food := range resp.Foods {
							score := 0.0
							if food.Score != nil {
								score = *food.Score
							}
							fdcMatches = append(fdcMatches, FDCMatch{
								FDCID:       food.FDCID,
								Description: food.Description,
								DataType:    food.DataType,
								Score:       score,
							})
						}
					}
				}
			}
			if fdcMatches == nil {
				fdcMatches = []FDCMatch{}
			}

			type InspectResult struct {
				Food          *food.Food               `json:"food"`
				Ingredients   []*ingredient.Ingredient `json:"ingredients"`
				IssueCount    int                      `json:"issue_count"`
				Issues        []string                 `json:"issues"`
				SuggestedName string                   `json:"suggested_name"`
				PrepNote      string                   `json:"prep_note,omitempty"`
				Alternatives  []string                 `json:"alternatives,omitempty"`
				FDCMatches    []FDCMatch               `json:"fdc_matches"`
			}

			result := InspectResult{
				Food:          f,
				Ingredients:   foodIngredients,
				IssueCount:    len(issues),
				Issues:        issueList,
				SuggestedName: norm.Canonical,
				PrepNote:      norm.PrepNote,
				Alternatives:  norm.Alternatives,
				FDCMatches:    fdcMatches,
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
			f, err := c.Foods().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get food: %w", err)
			}

			// Run detectors to find issues.
			reg := newAuditRegistry(ctx, c, cmd)
			detFood := detector.Food{
				Name:            f.Name,
				FDCID:           f.FDCID,
				PropertyTypeIDs: propertyTypeIDs(f),
			}
			issues := reg.CheckAll(detFood)

			issueCategories := make([]string, len(issues))
			for i, iss := range issues {
				issueCategories[i] = iss.Category
			}

			// Generate normalized name.
			norm := normalize.Normalize(f.Name)

			// Fetch affected ingredients.
			pageSize := cmd.Int("page-size")
			ingOpts := &ingredient.ListOptions{ListOptions: pagination.ListOptions{
				PageSize: pageSize,
				Extra:    map[string]string{"food": fmt.Sprintf("%d", id)}},
			}
			ingPage, err := c.Ingredients().List(ctx, ingOpts)
			type AffectedIngredient struct {
				ID         int    `json:"id"`
				RecipeID   int    `json:"recipe_id"`
				RecipeName string `json:"recipe_name"`
			}
			var affectedIngredients []AffectedIngredient
			if err == nil && ingPage != nil {
				for _, ing := range ingPage.Results {
					for _, ref := range ing.UsedInRecipes {
						affectedIngredients = append(affectedIngredients, AffectedIngredient{
							ID:         ing.ID,
							RecipeID:   ref.ID,
							RecipeName: ref.Name,
						})
					}
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
				IssuesResolved:      issueCategories,
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
		Description: "Applies normalized name change to food. Requires --yes flag. Use --dry-run to preview PATCH payload.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Print the PATCH payload without sending",
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
			f, err := c.Foods().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get food: %w", err)
			}

			norm := normalize.Normalize(f.Name)

			// Build a minimal PATCH payload — avoid sending zero-value fields
			// (shopping="", properties=null) which cause Tandoor 500 errors.
			payload := map[string]any{
				"name": norm.Canonical,
			}
			data, _ := json.MarshalIndent(payload, "", "  ")

			dryRun := cmd.Bool("dry-run")
			yes := cmd.Bool("yes")

			if dryRun {
				fmt.Fprintln(cmd.ErrWriter, "[dry-run] PATCH /api/food/", id, "/")
				fmt.Fprintln(cmd.ErrWriter, string(data))
				return nil
			}

			if !yes {
				return fmt.Errorf("--yes flag is required to apply the fix (use --dry-run to preview)")
			}

			var updated map[string]any
			if err := c.DoJSON(ctx, "PATCH", "api/food/"+fmt.Sprintf("%d/", id), payload, &updated); err != nil {
				return fmt.Errorf("patch food %d: %w", id, err)
			}

			fmt.Fprintf(cmd.ErrWriter, "[fix] food %d: %q -> %q\n", id, f.Name, updated["name"])
			return printJSON(updated)
		},
	}
}
