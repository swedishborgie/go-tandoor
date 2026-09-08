// cmd/tandoor-cli/cmd_audit_connectors.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/detector"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/urfave/cli/v3"
)

// auditConnectorsCommand returns the `audit connectors` subcommand.
func auditConnectorsCommand() *cli.Command {
	return &cli.Command{
		Name:  "connectors",
		Usage: "Analyze connector ingredient lines (+, -, &)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			pageSize := cmd.Int("page-size")
			reg := newAuditRegistry(ctx, c, cmd)

			type ConnectorEntry struct {
				FoodID        int      `json:"food_id"`
				FoodName      string   `json:"food_name"`
				Connector     string   `json:"connector"`
				CanonicalName string   `json:"canonical_name"`
				IngredientID  int      `json:"ingredient_id"`
				OriginalText  string   `json:"original_text"`
				RequiresSplit bool     `json:"requires_split"`
				RecipeIDs     []int    `json:"recipe_ids"`
				Issues        []string `json:"issues"`
			}

			var results []ConnectorEntry

			// Fetch all foods using the food service.
			pageNum := 1
			for {
				opts := &food.ListOptions{ListOptions: pagination.ListOptions{Page: pageNum, PageSize: pageSize}}
				page, err := c.Foods().List(ctx, opts)
				if err != nil {
					return fmt.Errorf("list foods page %d: %w", pageNum, err)
				}

				for _, f := range page.Results {
					detFood := detector.Food{
						Name:            f.Name,
						FDCID:           f.FDCID,
						PropertyTypeIDs: propertyTypeIDs(&f),
					}
					// Check only connector_start category.
					issues := reg.CheckCategories(detFood, []string{"connector_start"})
					if len(issues) == 0 {
						continue
					}

					details, _ := issues[0].Details.(map[string]string)
					connector := details["connector"]
					requiresSplit := connector == "&"

					norm := normalize.Normalize(f.Name)

					// Fetch ingredients for this food to get recipe info.
					ingOpts := &ingredient.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize, Extra: map[string]string{"food": fmt.Sprintf("%d", f.ID)}}}
					ingPage, err := c.Ingredients().List(ctx, ingOpts)
					var recipeIDs []int
					if err == nil && ingPage != nil {
						for _, ing := range ingPage.Results {
							for _, ref := range ing.UsedInRecipes {
								recipeIDs = append(recipeIDs, ref.ID)
							}
						}
					}

					issueList := make([]string, len(issues))
					for i, iss := range issues {
						issueList[i] = iss.Category
					}

					// Pick the first ingredient for display (if any).
					var ingID int
					var originalText string
					if ingPage != nil && len(ingPage.Results) > 0 {
						ingID = ingPage.Results[0].ID
						originalText = ingPage.Results[0].OriginalText
					}

					results = append(results, ConnectorEntry{
						FoodID:        f.ID,
						FoodName:      f.Name,
						Connector:     connector,
						CanonicalName: norm.Canonical,
						IngredientID:  ingID,
						OriginalText:  originalText,
						RequiresSplit: requiresSplit,
						RecipeIDs:     recipeIDs,
						Issues:        issueList,
					})
				}

				if !page.HasNext() {
					break
				}
				pageNum = page.NextPageNumber()
				if pageNum == 0 {
					break
				}
			}

			if results == nil {
				results = []ConnectorEntry{}
			}

			// Summary stats.
			autoFixable := 0
			manualSplit := 0
			for _, r := range results {
				if r.RequiresSplit {
					manualSplit++
				} else {
					autoFixable++
				}
			}

			fmt.Fprintf(cmd.ErrWriter, "connector foods: %d (auto-fixable: %d, requires split: %d)\n",
				len(results), autoFixable, manualSplit)

			return printJSON(results)
		},
	}
}
