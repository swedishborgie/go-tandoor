// cmd/tandoor-cli/cmd_properties.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
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

func resolvePropertyTargetFoodID(ctx context.Context, c *tandoor.Client, cmd *cli.Command) (int, error) {
	if cmd.IsSet("food-id") {
		return cmd.Int("food-id"), nil
	}
	if cmd.IsSet("ingredient-id") {
		ingredientID := cmd.Int("ingredient-id")
		ing, err := c.Ingredients().Get(ctx, ingredientID)
		if err != nil {
			return 0, fmt.Errorf("get ingredient %d: %w", ingredientID, err)
		}
		if ing.Food == nil {
			return 0, fmt.Errorf("ingredient %d has no food attached", ingredientID)
		}
		return ing.Food.ID, nil
	}
	return 0, fmt.Errorf("set --food-id or --ingredient-id")
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
				Usage: "Preview payload without sending",
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

			foodID, err := resolvePropertyTargetFoodID(ctx, c, cmd)
			if err != nil {
				printError(err)
				return err
			}

			propertyTypeID := cmd.Int("property-type-id")
			propertyAmount := cmd.Float("property-amount")

			if cmd.Bool("dry-run") {
				preview := map[string]any{
					"food_id":          foodID,
					"property_type_id": propertyTypeID,
					"property_amount":  propertyAmount,
					"action":           "attach-or-update",
				}
				fmt.Fprintf(cmd.ErrWriter, "[dry-run] properties attach preview:\n")
				return printJSON(preview)
			}

			// Fetch the authoritative property type. Tandoor's writable
			// nested serializer matches nested objects by id and overwrites
			// every field sent, so the payload must carry the real name,
			// unit, and fdc_id — never a fabricated or borrowed placeholder.
			propType, err := property.NewTypeService(c).Get(ctx, propertyTypeID)
			if err != nil {
				return fmt.Errorf("get property type %d: %w (create it first with property-types create)", propertyTypeID, err)
			}

			// Fetch the food to get its authoritative properties array
			// (the food endpoint returns the correct properties, unlike
			// the property list which can include stale/merged entries).
			var food map[string]any
			if err := c.DoJSON(ctx, "GET", "api/food/"+fmt.Sprintf("%d/", foodID), nil, &food); err != nil {
				return fmt.Errorf("get food %d: %w", foodID, err)
			}

			// Find existing property with matching type from the food's properties
			existingProps, _ := food["properties"].([]any)
			var targetPropID *int
			for _, p := range existingProps {
				propMap, ok := p.(map[string]any)
				if !ok {
					continue
				}
				pt, ok := propMap["property_type"].(map[string]any)
				if !ok {
					continue
				}
				ptID, ok := pt["id"].(float64)
				if !ok {
					continue
				}
				if int(ptID) == propertyTypeID {
					propID, ok := propMap["id"].(float64)
					if ok {
						id := int(propID)
						targetPropID = &id
					}
					break
				}
			}

			var result any
			if targetPropID != nil {
				// Tandoor requires property_type on PATCH (validates name != blank)
				payload := map[string]any{
					"property_amount": propertyAmount,
					"property_type":   propType,
				}
				result = make(map[string]any)
				if err := c.DoJSON(ctx, "PATCH", "api/property/"+fmt.Sprintf("%d/", *targetPropID), payload, &result); err != nil {
					return fmt.Errorf("patch property %d: %w", *targetPropID, err)
				}
				// Also set properties_food_amount/unit on the food so recipe
				// calculations can use the per-100g properties.
				foodPayload := map[string]any{
					"properties_food_amount": 100,
					"properties_food_unit":   17, // g
				}
				var foodResult map[string]any
				if err := c.DoJSON(ctx, "PATCH", "api/food/"+fmt.Sprintf("%d/", foodID), foodPayload, &foodResult); err != nil {
					// Non-fatal: property was still updated
					fmt.Fprintf(cmd.ErrWriter, "[warn] could not set properties_food_amount on food %d: %v\n", foodID, err)
				}
			} else {
				// POST /api/property/ ignores the food field, so we PATCH the
				// food directly with its properties array including the new one.
				newProp := map[string]any{
					"property_type":   propType,
					"property_amount": propertyAmount,
				}
				existingProps = append(existingProps, newProp)
				foodPayload := map[string]any{
					"id":                     foodID,
					"properties":             existingProps,
					"properties_food_amount": 100,
					"properties_food_unit":   17, // g
				}
				result = make(map[string]any)
				if err := c.DoJSON(ctx, "PATCH", "api/food/"+fmt.Sprintf("%d/", foodID), foodPayload, &result); err != nil {
					return fmt.Errorf("patch food %d with properties: %w", foodID, err)
				}
			}

			action := "updated"
			if targetPropID == nil {
				action = "created"
			}
			fmt.Fprintf(cmd.ErrWriter, "[attach] property %s type=%d amount=%.2f to food %d\n", action, propertyTypeID, propertyAmount, foodID)
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
