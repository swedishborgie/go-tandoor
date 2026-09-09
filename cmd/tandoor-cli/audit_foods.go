// cmd/tandoor-cli/cmd_audit_foods.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/auditfood"
	"github.com/swedishborgie/go-tandoor/detector"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/urfave/cli/v3"
)

// auditFoodsCommand returns the `audit foods` subcommand.
func auditFoodsCommand() *cli.Command {
	return &cli.Command{
		Name:  "foods",
		Usage: "Scan all foods for naming issues",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "category",
				Usage: "Filter to a specific issue category (e.g., quantity_prefix)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)

			category := cmd.String("category")
			reg := auditfood.NewRegistry(ctx, c, func(format string, args ...any) {
				fmt.Fprintf(cmd.ErrWriter, "[warn] "+format+"\n", args...)
			})
			pageSize := cmd.Int("page-size")

			type AuditResult struct {
				FoodID          int      `json:"food_id"`
				Name            string   `json:"name"`
				SuggestedName   string   `json:"suggested_name"`
				FDCID           *int     `json:"fdc_id"`
				IngredientCount int      `json:"ingredient_count"`
				RecipeIDs       []int    `json:"recipe_ids"`
				Issues          []string `json:"issues"`
			}

			// fetchIngredientsInfo retrieves ingredient count and recipe IDs for a food.
			fetchIngredientsInfo := func(foodID int) (int, []int) {
				opts := &ingredient.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize, Extra: map[string]string{"food": fmt.Sprintf("%d", foodID)}}}
				page, err := c.Ingredients().List(ctx, opts)
				if err != nil || page == nil {
					return 0, nil
				}
				var recipeIDs []int
				seen := make(map[int]struct{})
				for _, ing := range page.Results {
					for _, ref := range ing.UsedInRecipes {
						if _, ok := seen[ref.ID]; !ok {
							seen[ref.ID] = struct{}{}
							recipeIDs = append(recipeIDs, ref.ID)
						}
					}
				}
				if recipeIDs == nil {
					recipeIDs = []int{}
				}
				return len(page.Results), recipeIDs
			}

			var results []AuditResult

			// Manual pagination through all foods.
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
						PropertyTypeIDs: auditfood.PropertyTypeIDs(f.Properties),
					}
					var issues []detector.Issue
					if category != "" {
						issues = reg.CheckCategories(detFood, []string{category})
					} else {
						issues = reg.CheckAll(detFood)
					}

					if len(issues) == 0 {
						continue
					}

					issueList := make([]string, len(issues))
					for i, iss := range issues {
						issueList[i] = iss.Category
					}

					norm := normalize.Normalize(f.Name)

					ingCount, recipeIDs := fetchIngredientsInfo(f.ID)

					results = append(results, AuditResult{
						FoodID:          f.ID,
						Name:            f.Name,
						SuggestedName:   norm.Canonical,
						FDCID:           f.FDCID,
						IngredientCount: ingCount,
						RecipeIDs:       recipeIDs,
						Issues:          issueList,
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
				results = []AuditResult{}
			}

			totalCount, _ := countAllFoods(ctx, c)
			fmt.Fprintf(cmd.ErrWriter, "scanned %d foods, found %d with issues\n",
				totalCount, len(results))

			return printJSON(results)
		},
	}
}

// countAllFoods returns the total food count from the first page.
func countAllFoods(ctx context.Context, c *tandoor.Client) (int, error) {
	opts := &food.ListOptions{ListOptions: pagination.ListOptions{PageSize: 1}}
	page, err := c.Foods().List(ctx, opts)
	if err != nil {
		return -1, err
	}
	return page.Count, nil
}
