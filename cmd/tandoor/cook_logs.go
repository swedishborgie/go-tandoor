package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/misc"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/urfave/cli/v3"
)

func GetCookLogsCommand() *cli.Command {
	return &cli.Command{
		Name:  "cook-logs",
		Usage: "Cook log operations",
		Commands: []*cli.Command{
			cookLogsListCommand(),
			cookLogsCreateCommand(),
		},
	}
}

func cookLogsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List cook logs (GET /api/cook-log/)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			page, err := c.CookLogs().List(ctx, &misc.CookLogListOptions{ListOptions: pagination.ListOptions{PageSize: cmd.Int("page-size")}})
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func cookLogsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a cook log entry (POST /api/cook-log/)",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "recipe-id", Usage: "Recipe ID"},
			&cli.IntFlag{Name: "servings", Usage: "Servings cooked", Value: 1},
			&cli.IntFlag{Name: "rating", Usage: "Rating (optional)"},
			&cli.StringFlag{Name: "comment", Usage: "Comment"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			if !cmd.IsSet("recipe-id") {
				return fmt.Errorf("--recipe-id is required")
			}
			recipeID := cmd.Int("recipe-id")
			servings := cmd.Int("servings")
			comment := cmd.String("comment")
			payload := &misc.CookLog{
				Recipe:   &recipeID,
				Servings: &servings,
				Comment:  &comment,
			}
			if cmd.IsSet("rating") {
				rating := cmd.Int("rating")
				payload.Rating = &rating
			}
			created, err := c.CookLogs().Create(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] cook-log id=%d\n", created.ID)
			return printJSON(created)
		},
	}
}
