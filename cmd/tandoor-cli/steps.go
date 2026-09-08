// cmd/tandoor-cli/cmd_steps.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/step"
	"github.com/urfave/cli/v3"
)

// GetStepsCommand returns the top-level `steps` command group.
func GetStepsCommand() *cli.Command {
	return &cli.Command{
		Name:  "steps",
		Usage: "Step operations",
		Commands: []*cli.Command{
			stepsListCommand(),
			stepsGetCommand(),
			stepsCreateCommand(),
			stepsUpdateCommand(),
			stepsPatchCommand(),
			stepsDeleteCommand(),
		},
	}
}

func stepsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List steps (GET /api/step/)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			pageSize := cmd.Int("page-size")

			opts := &step.ListOptions{ListOptions: pagination.ListOptions{PageSize: pageSize}}
			page, err := c.Steps().List(ctx, opts)
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func stepsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a step by ID (GET /api/step/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			st, err := c.Steps().Get(ctx, id)
			if err != nil {
				return err
			}
			return printJSON(st)
		},
	}
}

func stepFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "instruction",
			Usage: "Step instruction text",
		},
		&cli.IntFlag{
			Name:  "time",
			Usage: "Step time in minutes",
		},
		&cli.IntFlag{
			Name:  "order",
			Usage: "Step order",
		},
		&cli.BoolFlag{
			Name:  "show-as-header",
			Usage: "Display step as header",
		},
		&cli.BoolFlag{
			Name:  "no-show-ingredients-table",
			Usage: "Hide step's ingredients table in recipe view (default: shown)",
		},
	}
}

func applyStepFlags(cmd *cli.Command, st *step.Step) {
	if cmd.IsSet("name") {
		st.Name = cmd.String("name")
	}
	if cmd.IsSet("instruction") {
		st.Instruction = cmd.String("instruction")
	}
	if cmd.IsSet("time") {
		st.Time = cmd.Int("time")
	}
	if cmd.IsSet("order") {
		st.Order = cmd.Int("order")
	}
	if cmd.IsSet("show-as-header") {
		st.ShowAsHeader = cmd.Bool("show-as-header")
	}
	if cmd.IsSet("no-show-ingredients-table") {
		st.ShowIngredientsTable = false
	}
}

func stepsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a step (POST /api/step/)",
		ArgsUsage: "name",
		Flags:     stepFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			name := cmd.Args().First()
			if name == "" {
				return fmt.Errorf("step name is required")
			}

			payload := &step.Step{Name: name, ShowIngredientsTable: true}
			applyStepFlags(cmd, payload)

			created, err := c.Steps().Create(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] step %q id=%d\n", created.Name, created.ID)
			return printJSON(created)
		},
	}
}

func stepsUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a step (PUT /api/step/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Step name",
			},
		}, stepFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			payload, err := c.Steps().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get step %d: %w", id, err)
			}
			applyStepFlags(cmd, payload)

			updated, err := c.Steps().Update(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] step %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func stepsPatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Patch a step (PATCH /api/step/<id>/)",
		ArgsUsage: "id",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Step name",
			},
		}, stepFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			payload, err := c.Steps().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get step %d: %w", id, err)
			}
			applyStepFlags(cmd, payload)

			updated, err := c.Steps().Patch(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[patch] step %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func stepsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a step (DELETE /api/step/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			if err := c.Steps().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] step %d\n", id)
			return nil
		},
	}
}
