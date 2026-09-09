package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/action"
	"github.com/swedishborgie/go-tandoor/importexport"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/urfave/cli/v3"
)

func GetImportsCommand() *cli.Command {
	return &cli.Command{
		Name:  "imports",
		Usage: "Recipe import operations",
		Commands: []*cli.Command{
			importsListCommand(),
			importsGetCommand(),
			importsImportCommand(),
			importsImportAllCommand(),
			importsDeleteCommand(),
		},
	}
}

func GetImportLogsCommand() *cli.Command {
	return &cli.Command{
		Name:  "import-logs",
		Usage: "Import log operations",
		Commands: []*cli.Command{
			importLogsListCommand(),
		},
	}
}

func GetRecipeFromSourceCommand() *cli.Command {
	return &cli.Command{
		Name:  "recipe-from-source",
		Usage: "Recipe import from source operations",
		Commands: []*cli.Command{
			recipeFromSourceCreateCommand(),
		},
	}
}

func GetShareLinksCommand() *cli.Command {
	return &cli.Command{
		Name:  "share-links",
		Usage: "Share link operations",
		Commands: []*cli.Command{
			shareLinksListCommand(),
			shareLinksCreateCommand(),
			shareLinksDeleteCommand(),
		},
	}
}

func importsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List recipe imports (GET /api/recipe-import/)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			page, err := c.RecipeImports().List(ctx, &importexport.RecipeImportListOptions{ListOptions: pagination.ListOptions{PageSize: cmd.Int("page-size")}})
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func importsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get an import by ID (GET /api/recipe-import/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			item, err := c.RecipeImports().Get(ctx, id)
			if err != nil {
				return err
			}
			return printJSON(item)
		},
	}
}

func importsImportCommand() *cli.Command {
	return &cli.Command{
		Name:      "import",
		Usage:     "Import a pending recipe (POST /api/recipe-import/<id>/import_recipe/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			result, err := c.RecipeImports().ImportRecipe(ctx, id)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[import] recipe import %d processed\n", id)
			return printJSON(result)
		},
	}
}

func importsImportAllCommand() *cli.Command {
	return &cli.Command{
		Name:  "import-all",
		Usage: "Import all pending recipes (POST /api/recipe-import/import_all/)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			result, err := c.RecipeImports().ImportAll(ctx)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[import-all] pending imports processed\n")
			return printJSON(result)
		},
	}
}

func importsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a pending import (DELETE /api/recipe-import/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			if err := c.RecipeImports().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] import %d\n", id)
			return nil
		},
	}
}

func importLogsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List import logs (GET /api/import-log/)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			page, err := c.ImportLogs().List(ctx, &importexport.ImportLogListOptions{ListOptions: pagination.ListOptions{PageSize: cmd.Int("page-size")}})
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func recipeFromSourceCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Import recipe from source URL or payload (POST /api/recipe-from-source/)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "url", Usage: "Source URL"},
			&cli.StringFlag{Name: "data", Usage: "Raw source data (HTML/JSON)"},
			&cli.IntFlag{Name: "bookmarklet-id", Usage: "Bookmarklet import ID"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			req := &action.RecipeFromSourceRequest{
				URL:  cmd.String("url"),
				Data: cmd.String("data"),
			}
			if cmd.IsSet("bookmarklet-id") {
				id := cmd.Int("bookmarklet-id")
				req.Bookmarklet = &id
			}
			if req.URL == "" && req.Data == "" && req.Bookmarklet == nil {
				return fmt.Errorf("set --url, --data, or --bookmarklet-id")
			}
			resp, err := c.RecipeFromSource().Import(ctx, req)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] recipe-from-source request submitted\n")
			return printJSON(resp)
		},
	}
}

func shareLinksListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List share links (GET /api/share-link/)",
		Action: func(ctx context.Context, _ *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			var page map[string]any
			if err := c.DoJSON(ctx, "GET", "api/share-link/", nil, &page); err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func shareLinksCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a share link for a recipe (GET /api/share-link/<recipe_id>)",
		ArgsUsage: "recipe_id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			recipeID, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			resp, err := c.ShareLinks().Create(ctx, recipeID)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] share link created for recipe %d\n", recipeID)
			return printJSON(resp)
		},
	}
}

func shareLinksDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a share link (DELETE /api/share-link/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			if err := c.DoJSON(ctx, "DELETE", fmt.Sprintf("api/share-link/%d/", id), nil, nil); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] share link %d\n", id)
			return nil
		},
	}
}
