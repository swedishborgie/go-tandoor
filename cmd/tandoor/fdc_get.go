// cmd/tandoor/cmd_fdc_get.go

package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// fdcGetCommand returns the `fdc get <fdc_id>` subcommand for fetching nutrient data.
func fdcGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Fetch FDC food details including nutrients (by FDC ID)",
		ArgsUsage: "fdc_id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fdcID, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}

			fdcClient, err := newFdcClient(cmd)
			if err != nil {
				return fmt.Errorf("fdc client: %w", err)
			}

			// Fetch all nutrients (filter to the 4 we need after retrieval)
			nutrients := []int{}
			food, err := fdcClient.GetFood(ctx, fdcID, "", nutrients)
			if err != nil {
				return fmt.Errorf("fdc get: %w", err)
			}

			type NutrientEntry struct {
				Number   int      `json:"number"`
				Name     string   `json:"name"`
				Amount   *float64 `json:"amount"`
				UnitName string   `json:"unit_name"`
			}
			type FoodDetail struct {
				FDCID       int             `json:"fdc_id"`
				Description string          `json:"description"`
				DataType    string          `json:"data_type"`
				Nutrients   []NutrientEntry `json:"nutrients"`
				Per100Gram  bool            `json:"per_100_gram"`
			}

			result := FoodDetail{
				FDCID:       food.FDCID,
				Description: food.Description,
				DataType:    food.DataType,
				Per100Gram:  true,
			}

			for _, n := range food.FoodNutrients {
				if n.Nutrient == nil || n.Amount == nil {
					continue
				}
				result.Nutrients = append(result.Nutrients, NutrientEntry{
					Number:   int(n.Nutrient.ID),
					Name:     n.Nutrient.Name,
					Amount:   n.Amount,
					UnitName: n.Nutrient.UnitName,
				})
			}

			if result.Nutrients == nil {
				result.Nutrients = []NutrientEntry{}
			}

			return printJSON(result)
		},
	}
}
