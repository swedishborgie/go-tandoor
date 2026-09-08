package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/swedishborgie/go-tandoor/recipebook"
	"github.com/urfave/cli/v3"
)

func GetBooksCommand() *cli.Command {
	return &cli.Command{
		Name:  "books",
		Usage: "Recipe book operations",
		Commands: []*cli.Command{
			booksListCommand(),
			booksGetCommand(),
			booksCreateCommand(),
			booksUpdateCommand(),
			booksDeleteCommand(),
		},
	}
}

func GetBookEntriesCommand() *cli.Command {
	return &cli.Command{
		Name:  "book-entries",
		Usage: "Recipe book entry operations",
		Commands: []*cli.Command{
			bookEntriesListCommand(),
			bookEntriesCreateCommand(),
			bookEntriesDeleteCommand(),
		},
	}
}

func booksListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List recipe books (GET /api/recipe-book/)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			page, err := c.RecipeBooks().List(ctx, &recipebook.ListOptions{ListOptions: pagination.ListOptions{PageSize: cmd.Int("page-size")}})
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func booksGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a recipe book by ID (GET /api/recipe-book/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			book, err := c.RecipeBooks().Get(ctx, id)
			if err != nil {
				return err
			}
			return printJSON(book)
		},
	}
}

func applyBookFlags(cmd *cli.Command, payload *recipebook.RecipeBook) error {
	if cmd.IsSet("name") {
		payload.Name = cmd.String("name")
	}
	if cmd.IsSet("description") {
		payload.Description = cmd.String("description")
	}
	if cmd.IsSet("order") {
		payload.Order = cmd.Int("order")
	}
	if cmd.IsSet("shared") {
		sharedIDs, err := parseIntCSV(cmd.String("shared"))
		if err != nil {
			return fmt.Errorf("parse --shared: %w", err)
		}
		sharedUsers := make([]recipe.User, 0, len(sharedIDs))
		for _, id := range sharedIDs {
			sharedUsers = append(sharedUsers, recipe.User{ID: id})
		}
		payload.Shared = sharedUsers
	}
	return nil
}

func bookFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "name", Usage: "Book name"},
		&cli.StringFlag{Name: "description", Usage: "Book description"},
		&cli.IntFlag{Name: "order", Usage: "Book order"},
		&cli.StringFlag{Name: "shared", Usage: "Comma-separated shared user IDs"},
	}
}

func booksCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a recipe book (POST /api/recipe-book/)",
		ArgsUsage: "name",
		Flags:     bookFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			name := cmd.Args().First()
			if name == "" && !cmd.IsSet("name") {
				return fmt.Errorf("book name is required")
			}
			payload := &recipebook.RecipeBook{Name: name}
			if err := applyBookFlags(cmd, payload); err != nil {
				return err
			}
			created, err := c.RecipeBooks().Create(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] book %q id=%d\n", created.Name, created.ID)
			return printJSON(created)
		},
	}
}

func booksUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a recipe book (PUT /api/recipe-book/<id>/)",
		ArgsUsage: "id",
		Flags:     bookFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			payload, err := c.RecipeBooks().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get book %d: %w", id, err)
			}
			if err := applyBookFlags(cmd, payload); err != nil {
				return err
			}
			updated, err := c.RecipeBooks().Update(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] book %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func booksDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a recipe book (DELETE /api/recipe-book/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			if err := c.RecipeBooks().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] book %d\n", id)
			return nil
		},
	}
}

func bookEntriesListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List recipe book entries (GET /api/recipe-book-entry/)",
		ArgsUsage: "[book_id]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			args := cmd.Args().Slice()
			bookID := 0
			var err error
			if len(args) > 0 && args[0] != "" {
				bookID, err = parseIDArg(args[0])
				if err != nil {
					return err
				}
			}
			params := url.Values{}
			params.Set("page_size", strconv.Itoa(cmd.Int("page-size")))
			if bookID > 0 {
				params.Set("book", strconv.Itoa(bookID))
			}
			var page pagination.Paginated[recipebook.Entry]
			if err := c.DoJSON(ctx, "GET", "api/recipe-book-entry/?"+params.Encode(), nil, &page); err != nil {
				return err
			}
			return printJSON(&page)
		},
	}
}

func bookEntriesCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a recipe book entry (POST /api/recipe-book-entry/)",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "book-id", Usage: "Recipe book ID"},
			&cli.IntFlag{Name: "recipe-id", Usage: "Recipe ID"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			if !cmd.IsSet("book-id") || !cmd.IsSet("recipe-id") {
				return fmt.Errorf("--book-id and --recipe-id are required")
			}
			svc := recipebook.NewEntryService(c)
			created, err := svc.Create(ctx, &recipebook.Entry{
				Book:   cmd.Int("book-id"),
				Recipe: cmd.Int("recipe-id"),
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] book-entry id=%d\n", created.ID)
			return printJSON(created)
		},
	}
}

func bookEntriesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a recipe book entry (DELETE /api/recipe-book-entry/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			svc := recipebook.NewEntryService(c)
			if err := svc.Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] book-entry %d\n", id)
			return nil
		},
	}
}
