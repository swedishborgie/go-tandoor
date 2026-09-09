// cmd/tandoor/cmd_properties.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/auditfood"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/urfave/cli/v3"
)

// GetPropertiesCommand returns the top-level `properties` command group.
func GetPropertiesCommand() *cli.Command {
	return &cli.Command{
		Name:  "properties",
		Usage: "Property operations",
		Commands: []*cli.Command{
			propertiesListCommand(),
			propertiesGetCommand(),
			propertiesCreateCommand(),
			propertiesAttachCommand(),
			propertiesUpdateCommand(),
			propertiesDeleteCommand(),
			propertiesTypesCommand(),
		},
	}
}

// propertiesTypesCommand returns the `properties types` subcommand group.
// It exposes the server-enumerated property type set — the authoritative
// list of nutrient types a food should carry.
func propertiesTypesCommand() *cli.Command {
	return &cli.Command{
		Name:  "types",
		Usage: "Property type operations (read-only)",
		Commands: []*cli.Command{
			propertiesTypesListCommand(),
			propertiesTypesGetCommand(),
		},
	}
}

func propertiesTypesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List property types (GET /api/property-type/)",
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
			svc := property.NewTypeService(c)
			pageSize := cmd.Int("page-size")

			opts := &property.TypeListOptions{ListOptions: pagination.ListOptions{
				PageSize: pageSize,
				Search:   cmd.String("search"),
			}}
			page, err := svc.List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[property.Type], error) {
					opts.Page = pageNum
					return svc.List(ctx, opts)
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

func propertiesTypesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a property type by ID (GET /api/property-type/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			pt, err := property.NewTypeService(c).Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(pt)
		},
	}
}

// GetPropertyTypesCommand returns the top-level `property-types` command group.
func GetPropertyTypesCommand() *cli.Command {
	return &cli.Command{
		Name:  "property-types",
		Usage: "Property type operations",
		Commands: []*cli.Command{
			propertyTypesListCommand(),
			propertyTypesCreateCommand(),
			propertyTypesUpdateCommand(),
			propertyTypesDeleteCommand(),
		},
	}
}

func propertiesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List properties (GET /api/property/)",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "food-id",
				Usage: "Filter properties for a food ID",
			},
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
			pageSize := cmd.Int("page-size")

			listOpts := &property.ListOptions{ListOptions: &pagination.ListOptions{
				PageSize: pageSize,
				Search:   cmd.String("search"),
			}}
			if cmd.IsSet("food-id") {
				foodID := cmd.Int("food-id")
				listOpts.FoodID = &foodID
			}

			page, err := c.Properties().List(ctx, listOpts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[property.Property], error) {
					listOpts.Page = pageNum
					return c.Properties().List(ctx, listOpts)
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

func propertiesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a property by ID (GET /api/property/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			p, err := c.Properties().Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(p)
		},
	}
}

func propertyValueFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "property-type-id",
			Usage: "Property type ID",
		},
		&cli.FloatFlag{
			Name:  "property-amount",
			Usage: "Property amount value",
		},
		&cli.BoolFlag{
			Name:  "clear-property-amount",
			Usage: "Clear property amount",
		},
	}
}

func applyPropertyFlags(cmd *cli.Command, p *property.Property) {
	if cmd.IsSet("property-type-id") {
		p.Type = property.Type{ID: cmd.Int("property-type-id")}
	}
	if cmd.IsSet("property-amount") {
		amount := cmd.Float("property-amount")
		p.PropertyAmount = &amount
	}
	if cmd.IsSet("clear-property-amount") && cmd.Bool("clear-property-amount") {
		p.PropertyAmount = nil
	}
}

func propertiesCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a property (POST /api/property/)",
		ArgsUsage: "",
		Flags:     propertyValueFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)

			payload := &property.Property{}
			applyPropertyFlags(cmd, payload)

			created, err := c.Properties().Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] property id=%d\n", created.ID)
			return printJSON(created)
		},
	}
}

func propertiesAttachCommand() *cli.Command {
	return &cli.Command{
		Name:        "attach",
		Usage:       "Attach a property to a food (creates or updates via /api/property/)",
		Description: "Attach a nutrient property to a food per 100g. Idempotent — updates existing property.",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "ingredient-id",
				Usage: "Ingredient ID (resolved to ingredient food)",
			},
			&cli.IntFlag{
				Name:  "food-id",
				Usage: "Food ID",
			},
			&cli.IntFlag{
				Name:  "property-type-id",
				Usage: "Property type ID",
			},
			&cli.FloatFlag{
				Name:  "property-amount",
				Usage: "Property amount value",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview the planned action without sending",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			if !cmd.IsSet("property-type-id") {
				return fmt.Errorf("--property-type-id is required")
			}
			if !cmd.IsSet("property-amount") {
				return fmt.Errorf("--property-amount is required")
			}

			result, err := auditfood.Attach(ctx, c, &auditfood.AttachOptions{
				FoodID:         cmd.Int("food-id"),
				IngredientID:   cmd.Int("ingredient-id"),
				PropertyTypeID: cmd.Int("property-type-id"),
				Amount:         cmd.Float("property-amount"),
			}, cmd.Bool("dry-run"))
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[attach] property %s (type=%d amount=%.2f)\n", result.Action, cmd.Int("property-type-id"), cmd.Float("property-amount"))
			return printJSON(result)
		},
	}
}

func propertiesUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a property (PUT /api/property/<id>/)",
		ArgsUsage: "id",
		Flags:     propertyValueFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := c.Properties().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get property %d: %w", id, err)
			}
			applyPropertyFlags(cmd, payload)

			updated, err := c.Properties().Update(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] property %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func propertiesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a property (DELETE /api/property/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			if err := c.Properties().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] property %d\n", id)
			return nil
		},
	}
}

func propertyTypeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "unit",
			Usage: "Property type unit (kcal, g, mg, etc.)",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Property type description",
		},
		&cli.IntFlag{
			Name:  "order",
			Usage: "Sort order",
		},
		&cli.StringFlag{
			Name:  "open-data-slug",
			Usage: "Open data slug",
		},
		&cli.IntFlag{
			Name:  "fdc-id",
			Usage: "USDA FDC ID",
		},
		&cli.BoolFlag{
			Name:  "clear-fdc-id",
			Usage: "Clear USDA FDC ID",
		},
	}
}

func applyPropertyTypeFlags(cmd *cli.Command, pt *property.Type) {
	if cmd.IsSet("name") {
		pt.Name = cmd.String("name")
	}
	if cmd.IsSet("unit") {
		pt.Unit = cmd.String("unit")
	}
	if cmd.IsSet("description") {
		pt.Description = cmd.String("description")
	}
	if cmd.IsSet("order") {
		pt.Order = cmd.Int("order")
	}
	if cmd.IsSet("open-data-slug") {
		pt.OpenDataSlug = cmd.String("open-data-slug")
	}
	if cmd.IsSet("fdc-id") {
		fdcID := cmd.Int("fdc-id")
		pt.FDCID = &fdcID
	}
	if cmd.IsSet("clear-fdc-id") && cmd.Bool("clear-fdc-id") {
		pt.FDCID = nil
	}
}

func propertyTypesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List property types (GET /api/property-type/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			svc := property.NewTypeService(c)
			pageSize := cmd.Int("page-size")

			opts := &property.TypeListOptions{ListOptions: pagination.ListOptions{
				PageSize: pageSize,
				Search:   cmd.String("search"),
			}}
			page, err := svc.List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			return outputWithJQ(ctx, page, cmd.String("jq"), cmd.String("output-file"))
		},
	}
}

func propertyTypesCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a property type (POST /api/property-type/)",
		ArgsUsage: "name",
		Flags:     propertyTypeFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			svc := property.NewTypeService(c)

			name := cmd.Args().First()
			if name == "" {
				return fmt.Errorf("property type name is required")
			}
			payload := &property.Type{Name: name}
			applyPropertyTypeFlags(cmd, payload)

			created, err := svc.Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] property-type %q id=%d\n", created.Name, created.ID)
			return printJSON(created)
		},
	}
}

func propertyTypesUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a property type (PUT /api/property-type/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Property type name",
			},
		}, propertyTypeFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			svc := property.NewTypeService(c)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := svc.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get property type %d: %w", id, err)
			}
			applyPropertyTypeFlags(cmd, payload)

			updated, err := svc.Update(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] property-type %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func propertyTypesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a property type (DELETE /api/property-type/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			svc := property.NewTypeService(c)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			if err := svc.Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] property-type %d\n", id)
			return nil
		},
	}
}
