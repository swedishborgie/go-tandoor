// cmd/tandoor/cmd_auth.go

package main

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/urfave/cli/v3"
)

// GetAuthCommand returns the top-level `auth` command group.
func GetAuthCommand() *cli.Command {
	return &cli.Command{
		Name:  "auth",
		Usage: "Authentication commands",
		Commands: []*cli.Command{
			authTokenCommand(),
			authOIDCCommand(),
			authListTokensCommand(),
		},
	}
}

func authTokenCommand() *cli.Command {
	return &cli.Command{
		Name:  "token",
		Usage: "Authenticate via POST /api-token-auth/",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "username",
				Aliases:  []string{"u"},
				Required: true,
				Usage:    "Tandoor username",
			},
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"p"},
				Required: true,
				Usage:    "Tandoor password",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			token, err := c.Auth().Auth(ctx, cmd.String("username"), cmd.String("password"))
			if err != nil {
				return err
			}
			fmt.Println(token.Token)
			return nil
		},
	}
}

func authOIDCCommand() *cli.Command {
	return &cli.Command{
		Name:  "oidc",
		Usage: "Authenticate via OIDC browser flow",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "oidc-backend",
				Usage: "OIDC backend slug (e.g. github, google-oauth2)",
				Value: "github",
			},
			&cli.StringFlag{
				Name:  "redirect-host",
				Usage: "Local callback host",
				Value: "localhost",
			},
			&cli.IntFlag{
				Name:  "redirect-port",
				Usage: "Local callback port",
				Value: 9999,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			token, err := oidcLogin(ctx, c,
				cmd.String("oidc-backend"),
				cmd.String("redirect-host"),
				cmd.Int("redirect-port"),
			)
			if err != nil {
				return err
			}
			fmt.Println(token)
			return nil
		},
	}
}

func authListTokensCommand() *cli.Command {
	return &cli.Command{
		Name:  "list-tokens",
		Usage: "List access tokens (GET /api/access-token/)",
		Action: func(ctx context.Context, _ *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			tokens, err := c.Auth().ListAccessTokens(ctx)
			if err != nil {
				return err
			}
			return printJSON(tokens)
		},
	}
}
