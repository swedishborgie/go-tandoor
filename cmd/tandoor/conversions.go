package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/unit"
	"github.com/urfave/cli/v3"
)

func GetConversionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "conversions",
		Usage: "Unit conversion operations",
		Commands: []*cli.Command{
			conversionsListCommand(),
			conversionsGetCommand(),
			conversionsCreateCommand(),
			conversionsCreateBatchCommand(),
			conversionsUpdateCommand(),
			conversionsPatchCommand(),
			conversionsDeleteCommand(),
		},
	}
}

func conversionFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "base-unit",
			Usage: "Base unit ID",
		},
		&cli.FloatFlag{
			Name:  "base-amount",
			Usage: "Base amount value",
		},
		&cli.IntFlag{
			Name:  "converted-unit",
			Usage: "Converted unit ID",
		},
		&cli.FloatFlag{
			Name:  "converted-amount",
			Usage: "Converted amount value",
		},
		&cli.IntFlag{
			Name:  "food",
			Usage: "Food ID for food-specific conversion",
		},
		&cli.BoolFlag{
			Name:  "clear-food",
			Usage: "Clear food ID to make conversion global",
		},
		&cli.StringFlag{
			Name:  "open-data-slug",
			Usage: "Open data slug",
		},
	}
}

func applyConversionFlags(cmd *cli.Command, conv *unit.Conversion) {
	if cmd.IsSet("base-unit") {
		conv.BaseUnit = &unit.Ref{ID: cmd.Int("base-unit")}
	}
	if cmd.IsSet("base-amount") {
		conv.BaseAmount = cmd.Float("base-amount")
	}
	if cmd.IsSet("converted-unit") {
		conv.ConvertedUnit = &unit.Ref{ID: cmd.Int("converted-unit")}
	}
	if cmd.IsSet("converted-amount") {
		conv.ConvertedAmount = cmd.Float("converted-amount")
	}
	if cmd.IsSet("food") {
		foodID := cmd.Int("food")
		conv.Food = &unit.FoodRef{ID: foodID}
	}
	if cmd.IsSet("clear-food") && cmd.Bool("clear-food") {
		conv.Food = nil
	}
	if cmd.IsSet("open-data-slug") {
		conv.OpenDataSlug = cmd.String("open-data-slug")
	}
}

func conversionsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List unit conversions (GET /api/unit-conversion/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
			&cli.IntFlag{
				Name:  "food-id",
				Usage: "Filter by food ID",
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Collect all pages into a single array",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			listOpts := &pagination.ListOptions{
				PageSize: cmd.Int("page-size"),
				Search:   cmd.String("search"),
			}
			opts := &unit.ConversionListOptions{ListOptions: listOpts}
			if cmd.IsSet("food-id") {
				foodID := cmd.Int("food-id")
				opts.FoodID = &foodID
			}
			page, err := c.UnitConversions().List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[unit.Conversion], error) {
					opts.Page = pageNum
					return c.UnitConversions().List(ctx, opts)
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

func conversionsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a unit conversion by ID (GET /api/unit-conversion/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			conv, err := c.UnitConversions().Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(conv)
		},
	}
}

func conversionsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:        "create",
		Usage:       "Create a unit conversion (POST /api/unit-conversion/)",
		Description: "Create a per-food unit conversion from base unit to grams. Required for macro calculations.",
		ArgsUsage:   "",
		Flags: append(conversionFlags(), &cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Preview payload without sending",
		}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			if !cmd.IsSet("base-unit") || !cmd.IsSet("base-amount") || !cmd.IsSet("converted-unit") || !cmd.IsSet("converted-amount") {
				return fmt.Errorf("--base-unit, --base-amount, --converted-unit, and --converted-amount are required")
			}

			payload := &unit.Conversion{}
			applyConversionFlags(cmd, payload)
			if cmd.Bool("dry-run") {
				fmt.Fprintf(cmd.ErrWriter, "[dry-run] POST api/unit-conversion/ payload:\n")
				return printJSON(payload)
			}
			created, err := c.UnitConversions().Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] conversion id=%d\n", created.ID)
			return printJSON(created)
		},
	}
}

type unitConversionBatchFile struct {
	Conversions []map[string]any `json:"conversions"`
}

// refID extracts an integer ID from either the shorthand form (27) or the
// object form ({"id": 27}) used in unit-conversion JSON.
func refID(item map[string]any, key string) *int {
	switch v := item[key].(type) {
	case float64:
		id := int(v)
		return &id
	case map[string]any:
		if id, ok := v["id"].(float64); ok {
			i := int(id)
			return &i
		}
	}
	return nil
}

func conversionsFromRaw(items []map[string]any) []unit.Conversion {
	var list []unit.Conversion
	for _, item := range items {
		conv := unit.Conversion{}
		if id := refID(item, "base_unit"); id != nil {
			conv.BaseUnit = &unit.Ref{ID: *id}
		}
		if v, ok := item["base_amount"].(float64); ok {
			conv.BaseAmount = v
		}
		if id := refID(item, "converted_unit"); id != nil {
			conv.ConvertedUnit = &unit.Ref{ID: *id}
		}
		if v, ok := item["converted_amount"].(float64); ok {
			conv.ConvertedAmount = v
		}
		if id := refID(item, "food"); id != nil {
			conv.Food = &unit.FoodRef{ID: *id}
		}
		list = append(list, conv)
	}
	return list
}

func loadUnitConversionsFromFile(path string) ([]unit.Conversion, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read conversions file %q: %w", path, err)
	}
	// Accept a bare array of conversions (integer IDs) or a {"conversions": [...]}
	// wrapper with the same item shape.
	var rawList []map[string]any
	if err := json.Unmarshal(body, &rawList); err == nil {
		return conversionsFromRaw(rawList), nil
	}
	var wrapped unitConversionBatchFile
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Conversions != nil {
		return conversionsFromRaw(wrapped.Conversions), nil
	}
	return nil, fmt.Errorf("parse conversions JSON %q: expected array or {\"conversions\": [...]}", path)
}

func conversionsCreateBatchCommand() *cli.Command {
	return &cli.Command{
		Name:  "create-batch",
		Usage: "Create multiple unit conversions from JSON file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Path to JSON file (array or {\"conversions\": [...]})",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			filePath := cmd.String("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			conversions, err := loadUnitConversionsFromFile(filePath)
			if err != nil {
				printError(err)
				return err
			}
			if len(conversions) == 0 {
				return fmt.Errorf("no conversions found in %q", filePath)
			}

			created := make([]unit.Conversion, 0, len(conversions))
			for i, conv := range conversions {
				if conv.BaseUnit == nil || conv.ConvertedUnit == nil || conv.BaseAmount == 0 || conv.ConvertedAmount == 0 {
					return fmt.Errorf("conversion at index %d is missing required fields", i)
				}
				newConv, err := c.UnitConversions().Create(ctx, &conv)
				if err != nil {
					return fmt.Errorf("create conversion at index %d: %w", i, err)
				}
				created = append(created, *newConv)
			}

			fmt.Fprintf(cmd.ErrWriter, "[create-batch] conversions created=%d\n", len(created))
			return printJSON(created)
		},
	}
}

func conversionsUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a unit conversion (PUT /api/unit-conversion/<id>/)",
		ArgsUsage: "id",
		Flags:     conversionFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			payload, err := c.UnitConversions().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get conversion %d: %w", id, err)
			}
			applyConversionFlags(cmd, payload)

			updated, err := c.UnitConversions().Update(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] conversion %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func conversionsPatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Patch a unit conversion (PATCH /api/unit-conversion/<id>/)",
		ArgsUsage: "id",
		Flags:     conversionFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			payload, err := c.UnitConversions().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get conversion %d: %w", id, err)
			}
			applyConversionFlags(cmd, payload)

			updated, err := c.UnitConversions().Patch(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[patch] conversion %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func conversionsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a unit conversion (DELETE /api/unit-conversion/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			if err := c.UnitConversions().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] conversion %d\n", id)
			return nil
		},
	}
}
