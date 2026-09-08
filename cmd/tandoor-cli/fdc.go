// cmd/tandoor-cli/cmd_fdc.go

package main

import (
	"fmt"
	"os"

	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/urfave/cli/v3"
)

// GetFdcCommand returns the top-level `fdc` command group.
func GetFdcCommand() *cli.Command {
	return &cli.Command{
		Name:  "fdc",
		Usage: "USDA FoodData Central API commands",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "fdc-api-key",
				Usage: "USDA FDC API key (defaults to project key)",
			},
		},
		Commands: []*cli.Command{
			fdcSearchCommand(),
			fdcFoodCommand(),
			fdcGetCommand(),
		},
	}
}

// newFdcClient creates an FDC API client from CLI flags or environment.
func newFdcClient(cmd *cli.Command) (*fdc.Client, error) {
	apiKey := cmd.String("fdc-api-key")
	if apiKey == "" {
		apiKey = os.Getenv("FDC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("FDC API key required (set --fdc-api-key or FDC_API_KEY env var)")
	}

	opts := []fdc.ClientOption{fdc.WithAPIKey(apiKey)}
	if baseURL := os.Getenv("FDC_BASE_URL"); baseURL != "" {
		opts = append(opts, fdc.WithBaseURL(baseURL))
	}
	return fdc.NewClient(opts...)
}
