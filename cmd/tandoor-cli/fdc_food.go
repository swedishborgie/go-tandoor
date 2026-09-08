// cmd/tandoor-cli/cmd_fdc_food.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/urfave/cli/v3"
)

// fdcFoodCommand returns the `fdc food <id> lookup` subcommand group.
func fdcFoodCommand() *cli.Command {
	return &cli.Command{
		Name:  "food",
		Usage: "Look up FDC ID for a Tandoor food",
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

			// Normalize the food name and search FDC.
			norm := normalize.Normalize(f.Name)

			type FDCMatch struct {
				FDCID       int      `json:"fdc_id"`
				Description string   `json:"description"`
				DataType    string   `json:"data_type"`
				Score       *float64 `json:"score,omitempty"`
			}
			type LookupResult struct {
				FoodID        int        `json:"food_id"`
				CurrentName   string     `json:"current_name"`
				CanonicalName string     `json:"canonical_name"`
				FDCMatches    []FDCMatch `json:"fdc_matches"`
			}

			result := LookupResult{
				FoodID:        id,
				CurrentName:   f.Name,
				CanonicalName: norm.Canonical,
			}

			// Search FDC using the canonical name.
			fdcClient, err := newFdcClient(cmd)
			if err != nil {
				return fmt.Errorf("fdc client: %w", err)
			}

			limit := 10

			resp, err := fdcClient.SearchFoods(ctx, norm.Canonical, nil, &limit, nil, "", "", "")
			if err != nil {
				return fmt.Errorf("fdc search: %w", err)
			}

			for _, food := range resp.Foods {
				result.FDCMatches = append(result.FDCMatches, FDCMatch{
					FDCID:       food.FDCID,
					Description: food.Description,
					DataType:    food.DataType,
					Score:       food.Score,
				})
			}

			if result.FDCMatches == nil {
				result.FDCMatches = []FDCMatch{}
			}

			fmt.Fprintf(cmd.ErrWriter, "total hits: %d, page: %d/%d\n",
				resp.TotalHits, resp.CurrentPage, resp.TotalPages)

			return printJSON(result)
		},
	}
}
