package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/shopping"
	"github.com/swedishborgie/go-tandoor/unit"
	"github.com/urfave/cli/v3"
)

func GetShoppingCommand() *cli.Command {
	return &cli.Command{
		Name:  "shopping",
		Usage: "Shopping list operations",
		Commands: []*cli.Command{
			shoppingListCommand(),
			shoppingGetCommand(),
			shoppingCreateCommand(),
			shoppingDeleteCommand(),
			shoppingEntriesCommand(),
			shoppingAddEntryCommand(),
			shoppingBulkUpdateCommand(),
			shoppingAddRecipeCommand(),
			shoppingCreateEntriesFromRecipeCommand(),
		},
	}
}

func shoppingListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List shopping lists (GET /api/shopping-list/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			opts := &shopping.ListListOptions{ListOptions: pagination.ListOptions{
				PageSize: cmd.Int("page-size"),
				Search:   cmd.String("search"),
			}}
			page, err := c.ShoppingLists().List(ctx, opts)
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func shoppingGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a shopping list by ID (GET /api/shopping-list/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			list, err := c.ShoppingLists().Get(ctx, id)
			if err != nil {
				return err
			}
			return printJSON(list)
		},
	}
}

func shoppingCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a shopping list (POST /api/shopping-list/)",
		ArgsUsage: "name",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "description", Usage: "List description"},
			&cli.StringFlag{Name: "color", Usage: "List color (hex)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			name := cmd.Args().First()
			if name == "" {
				return fmt.Errorf("shopping list name is required")
			}
			created, err := c.ShoppingLists().Create(ctx, &shopping.List{
				Name:        name,
				Description: cmd.String("description"),
				Color:       cmd.String("color"),
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] shopping list %q id=%d\n", created.Name, created.ID)
			return printJSON(created)
		},
	}
}

func shoppingDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a shopping list (DELETE /api/shopping-list/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			if err := c.ShoppingLists().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] shopping list %d\n", id)
			return nil
		},
	}
}

func shoppingEntriesCommand() *cli.Command {
	return &cli.Command{
		Name:      "entries",
		Usage:     "List entries in a shopping list (GET /api/shopping-list-entry/)",
		ArgsUsage: "shopping_list_id",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			listID, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			params := url.Values{}
			params.Set("shopping_list", strconv.Itoa(listID))
			params.Set("page_size", strconv.Itoa(cmd.Int("page-size")))
			if cmd.String("search") != "" {
				params.Set("search", cmd.String("search"))
			}

			var page pagination.Paginated[shopping.ListEntry]
			if err := c.DoJSON(ctx, "GET", "api/shopping-list-entry/?"+params.Encode(), nil, &page); err != nil {
				return err
			}
			return printJSON(&page)
		},
	}
}

func shoppingAddEntryCommand() *cli.Command {
	return &cli.Command{
		Name:  "add-entry",
		Usage: "Add an entry to a shopping list (POST /api/shopping-list-entry/)",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "list-id", Usage: "Shopping list ID"},
			&cli.IntFlag{Name: "food-id", Usage: "Food ID"},
			&cli.IntFlag{Name: "unit-id", Usage: "Unit ID"},
			&cli.FloatFlag{Name: "amount", Usage: "Amount"},
			&cli.StringFlag{Name: "note", Usage: "Entry note"},
			&cli.IntFlag{Name: "order", Usage: "Display order"},
			&cli.BoolFlag{Name: "is-header", Usage: "Mark as header"},
			&cli.BoolFlag{Name: "no-amount", Usage: "Hide amount"},
			&cli.StringFlag{Name: "original-text", Usage: "Original text"},
			&cli.BoolFlag{Name: "checked", Usage: "Mark checked"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			if !cmd.IsSet("list-id") {
				return fmt.Errorf("--list-id is required")
			}
			listID := cmd.Int("list-id")

			payload := &shopping.ListEntry{
				Lists: []shopping.List{{ID: listID}},
			}
			if cmd.IsSet("food-id") {
				foodID := cmd.Int("food-id")
				if f, err := c.Foods().Get(ctx, foodID); err == nil {
					payload.Food = &food.Shopping{ID: foodID, Name: f.Name}
				} else {
					payload.Food = &food.Shopping{ID: foodID}
				}
			}
			if cmd.IsSet("unit-id") {
				payload.Unit = &unit.Unit{ID: cmd.Int("unit-id")}
			}
			if cmd.IsSet("amount") {
				amount := cmd.Float("amount")
				payload.Amount = &amount
			}
			if cmd.IsSet("note") {
				payload.Note = cmd.String("note")
			}
			if cmd.IsSet("order") {
				payload.Order = cmd.Int("order")
			}
			if cmd.IsSet("is-header") {
				payload.IsHeader = cmd.Bool("is-header")
			}
			if cmd.IsSet("no-amount") {
				payload.NoAmount = cmd.Bool("no-amount")
			}
			if cmd.IsSet("original-text") {
				payload.OriginalText = cmd.String("original-text")
			}
			if cmd.IsSet("checked") {
				checked := cmd.Bool("checked")
				payload.Checked = &checked
			}

			created, err := c.ShoppingEntries().Create(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[add-entry] shopping entry id=%d\n", created.ID)
			return printJSON(created)
		},
	}
}

func shoppingBulkUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:  "bulk-update",
		Usage: "Bulk-update shopping entries (POST /api/shopping-list-entry/bulk/)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "ids", Usage: "Comma-separated shopping entry IDs"},
			&cli.BoolFlag{Name: "checked", Usage: "Set checked state"},
			&cli.StringFlag{Name: "shopping-lists-add", Usage: "Comma-separated list IDs to add"},
			&cli.StringFlag{Name: "shopping-lists-remove", Usage: "Comma-separated list IDs to remove"},
			&cli.StringFlag{Name: "shopping-lists-set", Usage: "Comma-separated list IDs to set"},
			&cli.BoolFlag{Name: "shopping-lists-remove-all", Usage: "Remove all shopping lists"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			ids, err := parseIntCSV(cmd.String("ids"))
			if err != nil {
				return fmt.Errorf("parse --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}
			add, err := parseIntCSV(cmd.String("shopping-lists-add"))
			if err != nil {
				return fmt.Errorf("parse --shopping-lists-add: %w", err)
			}
			remove, err := parseIntCSV(cmd.String("shopping-lists-remove"))
			if err != nil {
				return fmt.Errorf("parse --shopping-lists-remove: %w", err)
			}
			set, err := parseIntCSV(cmd.String("shopping-lists-set"))
			if err != nil {
				return fmt.Errorf("parse --shopping-lists-set: %w", err)
			}

			payload := &shopping.ListEntryBulk{
				IDs:            ids,
				ListsAdd:       add,
				ListsRemove:    remove,
				ListsSet:       set,
				ListsRemoveAll: cmd.Bool("shopping-lists-remove-all"),
			}
			if cmd.IsSet("checked") {
				checked := cmd.Bool("checked")
				payload.Checked = &checked
			}

			result, err := c.ShoppingEntries().BulkUpdate(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[bulk-update] shopping entries updated=%d\n", len(ids))
			return printJSON(result)
		},
	}
}

func shoppingAddRecipeCommand() *cli.Command {
	return &cli.Command{
		Name:      "add-recipe",
		Usage:     "Add a recipe to a shopping list (POST /api/shopping-list-recipe/)",
		ArgsUsage: "recipe_id",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "list-id", Usage: "Shopping list ID"},
			&cli.FloatFlag{Name: "servings", Usage: "Recipe servings override"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			recipeID, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			if !cmd.IsSet("list-id") {
				return fmt.Errorf("--list-id is required")
			}

			payload := map[string]any{
				"recipe":        recipeID,
				"shopping_list": cmd.Int("list-id"),
			}
			if cmd.IsSet("servings") {
				payload["servings"] = cmd.Float("servings")
			}

			var created map[string]any
			if err := c.DoJSON(ctx, "POST", "api/shopping-list-recipe/", payload, &created); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[add-recipe] recipe %d added to shopping list %d\n", recipeID, cmd.Int("list-id"))
			return printJSON(created)
		},
	}
}

func shoppingCreateEntriesFromRecipeCommand() *cli.Command {
	return &cli.Command{
		Name:      "create-entries-from-recipe",
		Usage:     "Create shopping entries from a shopping-list-recipe link",
		ArgsUsage: "shopping_list_recipe_id",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "shopping-list-ids", Usage: "Comma-separated shopping list IDs"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			listRecipeID, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			shoppingListIDs, err := parseIntCSV(cmd.String("shopping-list-ids"))
			if err != nil {
				return fmt.Errorf("parse --shopping-list-ids: %w", err)
			}
			result, err := c.ShoppingRecipes().BulkCreateEntries(ctx, listRecipeID, &shopping.ListEntryBulkCreate{
				ListIDs: shoppingListIDs,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create-entries-from-recipe] shopping-list-recipe %d\n", listRecipeID)
			return printJSON(result)
		},
	}
}
