// tandoor is a CLI tool for interacting with the Tandoor Recipes API.
//
// It supports API key authentication, OIDC browser login, and request/response
// recording via daytripper (HAR format).
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/swedishborgie/daytripper"
	"github.com/swedishborgie/daytripper/receiver"
	"github.com/urfave/cli/v3"

	"github.com/swedishborgie/go-tandoor"
)

// version is set at build time via -ldflags "-X main.version=...":
//   - goreleaser release builds: the git tag (e.g. v0.1.0)
//   - local builds: git describe output (tag + commit, "-dirty" when uncommitted)
//
// Builds without ldflags report "dev".
var version = "dev"

func main() {
	app := &cli.Command{
		Name:      "tandoor",
		Usage:     "CLI for the Tandoor Recipes API",
		Version:   version,
		UsageText: "tandoor [global options] <command> [command options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "base-url",
				Aliases:   []string{"b"},
				Usage:     "Tandoor instance base URL (required; or set TANDOOR_BASE_URL)",
				Sources:   cli.EnvVars("TANDOOR_BASE_URL"),
				TakesFile: false,
			},
			&cli.StringFlag{
				Name:      "token",
				Aliases:   []string{"t"},
				Usage:     "API access token (tda_...)",
				Sources:   cli.EnvVars("TANDOOR_TOKEN"),
				TakesFile: false,
			},
			&cli.StringFlag{
				Name:      "har-file",
				Usage:     "Record requests/responses to HAR file (e.g. output.har)",
				Sources:   cli.EnvVars("TANDOOR_HAR_FILE"),
				TakesFile: true,
			},
			&cli.IntFlag{
				Name:  "page-size",
				Usage: "Page size for list operations",
				Value: 50,
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			baseURL := cmd.String("base-url")
			if baseURL == "" {
				return ctx, fmt.Errorf("no base URL set: use --base-url or TANDOOR_BASE_URL")
			}
			opts := []tandoor.ClientOption{tandoor.WithBaseURL(baseURL)}

			// Set up daytripper recording if requested.
			harFile := cmd.String("har-file")
			if harFile != "" {
				r := receiver.NewHARFileReceiver(harFile)
				dt, err := daytripper.New(
					daytripper.WithReceiver(r),
					daytripper.WithCreator("tandoor"),
					daytripper.WithVersion(version),
				)
				if err != nil {
					return ctx, fmt.Errorf("daytripper: %w", err)
				}

				hc := &http.Client{
					Transport: dt,
					Timeout:   30 * time.Second,
				}
				ctx = context.WithValue(ctx, ctxKeyDT, dt)
				ctx = context.WithValue(ctx, ctxKeyHARFile, harFile)
				opts = append(opts, tandoor.WithHTTPClient(hc))
			}

			token := cmd.String("token")
			if token != "" {
				opts = append(opts, tandoor.WithAccessToken(token))
			}

			c, err := tandoor.NewClient(baseURL, opts...)
			if err != nil {
				return ctx, err
			}
			return context.WithValue(ctx, ctxKeyClient, c), nil
		},
		After: func(ctx context.Context, _ *cli.Command) error {
			// Flush daytripper if it was set up.
			if dt := ctx.Value(ctxKeyDT); dt != nil {
				dt := dt.(*daytripper.DayTripper)
				if err := dt.Flush(); err != nil {
					fmt.Fprintf(os.Stderr, "warn: daytripper flush: %v\n", err)
				}
				if err := dt.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "warn: daytripper close: %v\n", err)
				}
				harFile := ctx.Value(ctxKeyHARFile).(string)
				fmt.Fprintf(os.Stderr, "recorded to %s\n", harFile)
			}
			return nil
		},
		Commands: []*cli.Command{
			GetAuthCommand(),
			GetRecipesCommand(),
			GetShoppingCommand(),
			GetMealPlansCommand(),
			GetMealTypesCommand(),
			GetCookLogsCommand(),
			GetBooksCommand(),
			GetBookEntriesCommand(),
			GetImportsCommand(),
			GetImportLogsCommand(),
			GetRecipeFromSourceCommand(),
			GetShareLinksCommand(),
			GetFoodsCommand(),
			GetKeywordsCommand(),
			GetUnitsCommand(),
			GetConversionsCommand(),
			GetIngredientsCommand(),
			GetStepsCommand(),
			GetPropertiesCommand(),
			GetPropertyTypesCommand(),
			GetAuditCommand(),
			GetFdcCommand(),
			GetMCPCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Context key types for storing values in context.
type ctxKey int

const (
	ctxKeyClient ctxKey = iota
	ctxKeyDT
	ctxKeyHARFile
)
