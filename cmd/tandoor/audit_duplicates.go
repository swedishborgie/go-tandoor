// cmd/tandoor/cmd_audit_duplicates.go

package main

import (
	"context"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/auditfood"
	"github.com/urfave/cli/v3"
)

// auditDuplicatesCommand returns the `audit duplicates` subcommand.
func auditDuplicatesCommand() *cli.Command {
	return &cli.Command{
		Name:        "duplicates",
		Usage:       "Find groups of similar food names",
		Description: "Scans all foods, normalizes names, and reports groups above the similarity threshold (union-find over Jaccard word similarity).",
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:  "threshold",
				Usage: "Jaccard similarity threshold",
				Value: auditfood.DefaultDuplicateThreshold,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)

			result, err := auditfood.FindDuplicates(ctx, c, cmd.Float64("threshold"))
			if err != nil {
				return err
			}
			return printJSON(result)
		},
	}
}
