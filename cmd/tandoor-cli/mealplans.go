package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/mealplan"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/urfave/cli/v3"
)

func GetMealPlansCommand() *cli.Command {
	return &cli.Command{
		Name:  "mealplans",
		Usage: "Meal plan operations",
		Commands: []*cli.Command{
			mealPlansListCommand(),
			mealPlansGetCommand(),
			mealPlansCreateCommand(),
			mealPlansUpdateCommand(),
			mealPlansDeleteCommand(),
			mealPlansICALCommand(),
			mealPlansAutoPlanCommand(),
		},
	}
}

func GetMealTypesCommand() *cli.Command {
	return &cli.Command{
		Name:  "mealtypes",
		Usage: "Meal type operations",
		Commands: []*cli.Command{
			mealTypesListCommand(),
			mealTypesGetCommand(),
		},
	}
}

func mealPlansListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List meal plans (GET /api/meal-plan/)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "search", Aliases: []string{"q"}, Usage: "Search query"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			opts := &mealplan.ListOptions{ListOptions: pagination.ListOptions{PageSize: cmd.Int("page-size"), Search: cmd.String("search")}}
			page, err := c.MealPlans().List(ctx, opts)
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func mealPlansGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a meal plan by ID (GET /api/meal-plan/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			item, err := c.MealPlans().Get(ctx, id)
			if err != nil {
				return err
			}
			return printJSON(item)
		},
	}
}

func mealPlanUpsertFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "title", Usage: "Meal plan title"},
		&cli.IntFlag{Name: "recipe-id", Usage: "Recipe ID"},
		&cli.IntFlag{Name: "meal-type-id", Usage: "Meal type ID"},
		&cli.StringFlag{Name: "note", Usage: "Meal plan note"},
		&cli.FloatFlag{Name: "servings", Usage: "Servings override"},
		&cli.BoolFlag{Name: "shopping", Usage: "Mark as in shopping list"},
		&cli.BoolFlag{Name: "add-shopping", Usage: "Add to shopping list"},
		&cli.StringFlag{Name: "from-date", Usage: "From date/time (RFC3339 or YYYY-MM-DD)"},
		&cli.StringFlag{Name: "to-date", Usage: "To date/time (RFC3339 or YYYY-MM-DD)"},
	}
}

func applyMealPlanFlags(cmd *cli.Command, payload *mealplan.MealPlan) error {
	if cmd.IsSet("title") {
		payload.Title = cmd.String("title")
	}
	if cmd.IsSet("recipe-id") {
		payload.Recipe = &recipe.Overview{ID: cmd.Int("recipe-id")}
	}
	if cmd.IsSet("meal-type-id") {
		payload.MealType = &mealplan.MealType{ID: cmd.Int("meal-type-id")}
	}
	if cmd.IsSet("note") {
		payload.Note = cmd.String("note")
	}
	if cmd.IsSet("servings") {
		payload.Servings = cmd.Float("servings")
	}
	if cmd.IsSet("shopping") {
		payload.Shopping = cmd.Bool("shopping")
	}
	if cmd.IsSet("add-shopping") {
		payload.AddShopping = cmd.Bool("add-shopping")
	}
	if cmd.IsSet("from-date") {
		fromDate, err := parseDateOrDateTime(cmd.String("from-date"))
		if err != nil {
			return fmt.Errorf("parse --from-date: %w", err)
		}
		payload.FromDate = fromDate
	}
	if cmd.IsSet("to-date") {
		toDate, err := parseDateOrDateTime(cmd.String("to-date"))
		if err != nil {
			return fmt.Errorf("parse --to-date: %w", err)
		}
		payload.ToDate = toDate
	}
	return nil
}

func mealPlansCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a meal plan entry (POST /api/meal-plan/)",
		ArgsUsage: "",
		Flags:     mealPlanUpsertFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			payload := &mealplan.MealPlan{}
			if err := applyMealPlanFlags(cmd, payload); err != nil {
				return err
			}
			if payload.Recipe == nil && payload.Title == "" {
				return fmt.Errorf("set --recipe-id or --title")
			}
			created, err := c.MealPlans().Create(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[create] meal plan id=%d\n", created.ID)
			return printJSON(created)
		},
	}
}

func mealPlansUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update a meal plan entry (PUT /api/meal-plan/<id>/)",
		ArgsUsage: "id",
		Flags:     mealPlanUpsertFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			payload, err := c.MealPlans().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("get meal plan %d: %w", id, err)
			}
			if err := applyMealPlanFlags(cmd, payload); err != nil {
				return err
			}
			updated, err := c.MealPlans().Update(ctx, payload)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[update] meal plan %d updated\n", id)
			return printJSON(updated)
		},
	}
}

func mealPlansDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a meal plan entry (DELETE /api/meal-plan/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			if err := c.MealPlans().Delete(ctx, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[delete] meal plan %d\n", id)
			return nil
		},
	}
}

func mealPlansICALCommand() *cli.Command {
	return &cli.Command{
		Name:  "ical",
		Usage: "Get meal plan iCal feed (GET /api/meal-plan/ical/)",
		Action: func(ctx context.Context, _ *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			ical, err := c.MealPlans().ICAL(ctx, nil)
			if err != nil {
				return err
			}
			fmt.Println(ical)
			return nil
		},
	}
}

func mealPlansAutoPlanCommand() *cli.Command {
	return &cli.Command{
		Name:  "auto-plan",
		Usage: "Auto-generate meal plans (POST /api/auto-meal-plan/)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "start-date", Usage: "Start date/time (RFC3339 or YYYY-MM-DD)"},
			&cli.StringFlag{Name: "end-date", Usage: "End date/time (RFC3339 or YYYY-MM-DD)"},
			&cli.IntFlag{Name: "meal-type-id", Usage: "Meal type ID"},
			&cli.StringFlag{Name: "keywords", Usage: "Comma-separated keyword IDs"},
			&cli.StringFlag{Name: "keyword-mode", Usage: "Keyword mode (or|and)", Value: "or"},
			&cli.FloatFlag{Name: "servings", Usage: "Servings", Value: 1},
			&cli.StringFlag{Name: "shared", Usage: "Comma-separated shared user IDs"},
			&cli.BoolFlag{Name: "add-shopping", Usage: "Add generated meal plans to shopping"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			if !cmd.IsSet("start-date") || !cmd.IsSet("end-date") || !cmd.IsSet("meal-type-id") {
				return fmt.Errorf("--start-date, --end-date, and --meal-type-id are required")
			}
			startDate, err := parseDateOrDateTime(cmd.String("start-date"))
			if err != nil {
				return fmt.Errorf("parse --start-date: %w", err)
			}
			endDate, err := parseDateOrDateTime(cmd.String("end-date"))
			if err != nil {
				return fmt.Errorf("parse --end-date: %w", err)
			}
			keywords, err := parseIntCSV(cmd.String("keywords"))
			if err != nil {
				return fmt.Errorf("parse --keywords: %w", err)
			}
			shared, err := parseIntCSV(cmd.String("shared"))
			if err != nil {
				return fmt.Errorf("parse --shared: %w", err)
			}

			result, err := c.AutoPlan().Plan(ctx, &mealplan.AutoMealPlanRequest{
				StartDate:   startDate,
				EndDate:     endDate,
				MealTypeID:  cmd.Int("meal-type-id"),
				Keywords:    keywords,
				KeywordMode: cmd.String("keyword-mode"),
				Servings:    cmd.Float("servings"),
				Shared:      shared,
				AddShopping: cmd.Bool("add-shopping"),
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrWriter, "[auto-plan] generated meal plan request submitted\n")
			return printJSON(result)
		},
	}
}

func mealTypesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List meal types (GET /api/meal-type/)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			page, err := c.MealTypes().List(ctx, &pagination.ListOptions{PageSize: cmd.Int("page-size")})
			if err != nil {
				return err
			}
			return printJSON(page)
		},
	}
}

func mealTypesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a meal type by ID (GET /api/meal-type/<id>/)",
		ArgsUsage: "id",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			id, err := parseIDArg(cmd.Args().First())
			if err != nil {
				return err
			}
			item, err := c.MealTypes().Get(ctx, id)
			if err != nil {
				return err
			}
			return printJSON(item)
		},
	}
}
