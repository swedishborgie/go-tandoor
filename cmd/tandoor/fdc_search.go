package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/urfave/cli/v3"
)

// fdcSearchCommand returns the `fdc search <query>` subcommand.
func fdcSearchCommand() *cli.Command {
	return &cli.Command{
		Name:        "search",
		Usage:       "Search USDA FoodData Central by keyword",
		Description: "Search USDA FoodData Central by keyword. Returns FDC ID, description and data type for matching foods.",
		ArgsUsage:   "query",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "query",
				Usage: "Query strings to search (batch)",
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "Max results to return",
				Value: 5,
			},
			&cli.StringFlag{
				Name:  "type",
				Usage: "Filter by data type: Branded, Foundation, Survey (FNDDS), SR Legacy",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			queries := cmd.StringSlice("query")
			if len(queries) == 0 {
				if cmd.Args().First() != "" {
					queries = []string{cmd.Args().First()}
				}
			}
			if len(queries) == 0 {
				return fmt.Errorf("search query is required")
			}

			limit := cmd.Int("limit")
			pageSize := limit
			if pageSize > 200 {
				pageSize = 200
			}

			client, err := newFdcClient(cmd)
			if err != nil {
				return err
			}

			dataType, err := parseDataType(cmd.String("type"))
			if err != nil {
				return err
			}

			type SearchResult struct {
				FDCID       int     `json:"fdc_id"`
				Description string  `json:"description"`
				DataType    string  `json:"data_type"`
				Score       float64 `json:"score,omitempty"`
			}

			allResults := make(map[string][]SearchResult)
			for _, query := range queries {
				nilInt := (*int)(nil)
				resp, err := client.SearchFoods(ctx, query, dataType, &pageSize, nilInt, "", "", "")
				if err != nil {
					return fmt.Errorf("fdc search %q: %w", query, err)
				}
				foods := resp.Foods
				if len(foods) > limit {
					foods = foods[:limit]
				}
				results := make([]SearchResult, 0, len(foods))
				for _, f := range foods {
					r := SearchResult{
						FDCID:       f.FDCID,
						Description: f.Description,
						DataType:    f.DataType,
					}
					if f.Score != nil {
						r.Score = *f.Score
					}
					results = append(results, r)
				}
				allResults[query] = results
				fmt.Fprintf(cmd.ErrWriter, "query %q: total hits %d\n", query, resp.TotalHits)
			}

			return printJSON(allResults)
		},
	}
}

// parseDataType converts a string to fdc.DataType or returns nil for empty.
func parseDataType(s string) ([]fdc.DataType, error) {
	if s == "" {
		return nil, nil
	}

	dataTypeMap := map[string]fdc.DataType{
		"Branded":    fdc.DataTypeBranded,
		"Foundation": fdc.DataTypeFoundation,
		"Survey":     fdc.DataTypeSurvey,
		"SR Legacy":  fdc.DataTypeSRLegacy,
	}

	dt, ok := dataTypeMap[s]
	if !ok {
		return nil, fmt.Errorf("unknown data type %q (valid: Branded, Foundation, Survey (FNDDS), SR Legacy)", s)
	}
	return []fdc.DataType{dt}, nil
}
