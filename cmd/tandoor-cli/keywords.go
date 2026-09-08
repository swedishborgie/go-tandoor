// cmd/tandoor-cli/cmd_keywords.go

package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/keyword"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/urfave/cli/v3"
)

// GetKeywordsCommand returns the top-level `keywords` command group.
func GetKeywordsCommand() *cli.Command {
	return &cli.Command{
		Name:  "keywords",
		Usage: "Keyword operations",
		Commands: []*cli.Command{
			keywordsListCommand(),
			keywordsGetCommand(),
			keywordsCreateCommand(),
			keywordsUpdateCommand(),
			keywordsPatchCommand(),
			keywordsDeleteCommand(),
			keywordsMergeCommand(),
		},
	}
}

func keywordsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List keywords (GET /api/keyword/)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"q"},
				Usage:   "Search query",
			},
			&cli.IntFlag{
				Name:  "parent-id",
				Usage: "Filter by parent keyword ID",
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

			opts := &keyword.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize}}
			if search != "" {
				opts.Search = search
			}
			if cmd.IsSet("parent-id") {
				if opts.Extra == nil {
					opts.Extra = make(map[string]string)
				}
				opts.Extra["parent"] = strconv.Itoa(cmd.Int("parent-id"))
			}

			page, err := c.Keywords().List(ctx, opts)
			if err != nil {
				printError(err)
				return err
			}
			if cmd.Bool("all") {
				all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[keyword.Keyword], error) {
					opts.Page = pageNum
					return c.Keywords().List(ctx, opts)
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

func keywordsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a keyword by ID (GET /api/keyword/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			k, err := c.Keywords().Get(ctx, id)
			if err != nil {
				printError(err)
				return err
			}
			return printJSON(k)
		},
	}
}

func keywordFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "label",
			Usage: "Keyword label",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Keyword description",
		},
		&cli.IntFlag{
			Name:  "parent",
			Usage: "Parent keyword ID",
		},
	}
}

func keywordsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a keyword (POST /api/keyword/)",
		ArgsUsage: "name",
		Flags:     keywordFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			name := cmd.Args().First()
			if name == "" {
				return fmt.Errorf("keyword name is required")
			}

			payload := &keyword.Keyword{
				Name: name,
			}
			if cmd.IsSet("label") {
				payload.Label = cmd.String("label")
			}
			if cmd.IsSet("description") {
				payload.Description = cmd.String("description")
			}
			if cmd.IsSet("parent") {
				payload.Parent = cmd.Int("parent")
			}

			created, err := c.Keywords().Create(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] keyword %q id=%d\n", created.Name, created.ID)
			return printJSON(created)
		},
	}
}

func keywordsUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a keyword (PUT /api/keyword/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Keyword name",
			},
		}, keywordFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := c.Keywords().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get keyword %d: %w", id, err)
			}

			if cmd.IsSet("name") {
				payload.Name = cmd.String("name")
			}
			if cmd.IsSet("label") {
				payload.Label = cmd.String("label")
			}
			if cmd.IsSet("description") {
				payload.Description = cmd.String("description")
			}
			if cmd.IsSet("parent") {
				payload.Parent = cmd.Int("parent")
			}

			updated, err := c.Keywords().Update(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] keyword %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func keywordsPatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Patch a keyword (PATCH /api/keyword/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Keyword name",
			},
		}, keywordFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}

			payload, err := c.Keywords().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get keyword %d: %w", id, err)
			}

			if cmd.IsSet("name") {
				payload.Name = cmd.String("name")
			}
			if cmd.IsSet("label") {
				payload.Label = cmd.String("label")
			}
			if cmd.IsSet("description") {
				payload.Description = cmd.String("description")
			}
			if cmd.IsSet("parent") {
				payload.Parent = cmd.Int("parent")
			}

			updated, err := c.Keywords().Patch(ctx, payload)
			if err != nil {
				printError(err)
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[patch] keyword %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func keywordsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a keyword (DELETE /api/keyword/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				printError(err)
				return err
			}
			if err := c.Keywords().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] keyword %d\n", id)
			return nil
		},
	}
}

func keywordsMergeCommand() *cli.Command {
	return &cli.Command{
		Name:      "merge",
		Usage:     "Merge one keyword into another (PUT /api/keyword/<source>/merge/<target>/)",
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

			merged, err := c.Keywords().Merge(ctx, sourceID, targetID)
			if err != nil {
				return fmt.Errorf("merge keyword %d into %d: %w", sourceID, targetID, err)
			}
			fmt.Fprintf(cmd.ErrWriter, "[merge] keyword %d -> %d\n", sourceID, targetID)
			return printJSON(merged)
		},
	}
}
