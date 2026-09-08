// cmd/tandoor-cli/cmd_units.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/unit"
	"github.com/urfave/cli/v3"
)

// GetUnitsCommand returns the top-level `units` command group.
func GetUnitsCommand() *cli.Command {
	return &cli.Command{
		Name:  "units",
		Usage: "Unit operations",
		Commands: []*cli.Command{
			unitsListCommand(),
			unitsGetCommand(),
			unitsCreateCommand(),
			unitsUpdateCommand(),
			unitsPatchCommand(),
			unitsDeleteCommand(),
			unitsMergeCommand(),
		},
	}
}

func unitsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List units (GET /api/unit/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Collect all pages into a single array",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			search := cmd.String("search")
			pageSize := cmd.Int("page-size")

			opts := &unit.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize}}
			if search != "" {
				opts.Search = search
			}

			page, err := c.Units().List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[unit.Unit], error) {
					opts.Page = pageNum
					return c.Units().List(ctx, opts)
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

func unitsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a unit by ID (GET /api/unit/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			u, err := c.Units().Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(u)
		},
	}
}

func unitSharedFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "plural-name",
			Usage: "Plural display name",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Unit description",
		},
		&cli.StringFlag{
			Name:  "base-unit",
			Usage: "Base unit kind (weight, volume, etc.)",
		},
		&cli.StringFlag{
			Name:  "open-data-slug",
			Usage: "Open data slug",
		},
	}
}

func unitsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a unit (POST /api/unit/)",
		ArgsUsage: "name",
		Flags:     unitSharedFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			name := cmd.Args().First()
			if name == "" {
				return fmt.Errorf("unit name is required")
			}

			payload := &unit.Unit{Name: name}
			if cmd.IsSet("plural-name") {
				payload.PluralName = cmd.String("plural-name")
			}
			if cmd.IsSet("description") {
				payload.Description = cmd.String("description")
			}
			if cmd.IsSet("base-unit") {
				payload.BaseUnit = cmd.String("base-unit")
			}
			if cmd.IsSet("open-data-slug") {
				payload.OpenDataSlug = cmd.String("open-data-slug")
			}

			created, err := c.Units().Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] unit %q id=%d\n", created.Name, created.ID)
			return printJSON(created)
		},
	}
}

func unitsUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a unit (PUT /api/unit/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Unit name",
			},
		}, unitSharedFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := c.Units().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get unit %d: %w", id, err)
			}

			if cmd.IsSet("name") {
				payload.Name = cmd.String("name")
			}
			if cmd.IsSet("plural-name") {
				payload.PluralName = cmd.String("plural-name")
			}
			if cmd.IsSet("description") {
				payload.Description = cmd.String("description")
			}
			if cmd.IsSet("base-unit") {
				payload.BaseUnit = cmd.String("base-unit")
			}
			if cmd.IsSet("open-data-slug") {
				payload.OpenDataSlug = cmd.String("open-data-slug")
			}

			updated, err := c.Units().Update(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] unit %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func unitsPatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Patch a unit (PATCH /api/unit/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Unit name",
			},
		}, unitSharedFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := c.Units().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get unit %d: %w", id, err)
			}

			if cmd.IsSet("name") {
				payload.Name = cmd.String("name")
			}
			if cmd.IsSet("plural-name") {
				payload.PluralName = cmd.String("plural-name")
			}
			if cmd.IsSet("description") {
				payload.Description = cmd.String("description")
			}
			if cmd.IsSet("base-unit") {
				payload.BaseUnit = cmd.String("base-unit")
			}
			if cmd.IsSet("open-data-slug") {
				payload.OpenDataSlug = cmd.String("open-data-slug")
			}

			updated, err := c.Units().Patch(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[patch] unit %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func unitsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a unit (DELETE /api/unit/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			if err := c.Units().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] unit %d\n", id)
			return nil
		},
	}
}

func unitsMergeCommand() *cli.Command {
	return &cli.Command{
		Name:      "merge",
		Usage:     "Merge one unit into another (PUT /api/unit/<source>/merge/<target>/)",
		ArgsUsage: "source_id target_id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			args := cmd.Args().Slice()
			if len(args) < 2 {
				return fmt.Errorf("source_id and target_id are required")
			}

			sourceID, err := parseIDArg(args[0])
			if err != nil {
				return fmt.Errorf("invalid source_id: %w", err)
			}
			targetID, err := parseIDArg(args[1])
			if err != nil {
				return fmt.Errorf("invalid target_id: %w", err)
			}

			merged, err := c.Units().Merge(ctx, sourceID, targetID)
			if err != nil {
				return fmt.Errorf("merge unit %d into %d: %w", sourceID, targetID, err)
			}
			fmt.Fprintf(cmd.ErrWriter, "[merge] unit %d -> %d\n", sourceID, targetID)
			return printJSON(merged)
		},
	}
}
