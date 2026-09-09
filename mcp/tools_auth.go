package mcp

import (
	"context"
	"fmt"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/auth"
)

func registerAuthTools(d *deps) []toolDef {
	listTokens := mcpgo.NewTool("auth_list_tokens",
		mcpgo.WithDescription("List API access tokens for the current user (with scopes and expiry)."),
		jqParam(),
	)
	listTokensHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.Auth().ListAccessTokens(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	getToken := mcpgo.NewTool("auth_get_token",
		mcpgo.WithDescription("Get a single API access token by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Token ID")),
		jqParam(),
	)
	getTokenHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Auth().GetAccessToken(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	createToken := mcpgo.NewTool("auth_create_token",
		mcpgo.WithDescription("Create a new API access token for the current user. The token value is only returned in full immediately after creation."),
		mcpgo.WithString("scope", mcpgo.Description("Token scope (default \"read write\")")),
		mcpgo.WithString("expires", mcpgo.Required(), mcpgo.Description("Expiry timestamp (RFC3339)")),
		dryRunParam(),
	)
	createTokenHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		expiresStr, err := req.RequireString("expires")
		if err != nil {
			return errResult(fmt.Errorf("expires is required (RFC3339)")), nil
		}
		expires, err := time.Parse(time.RFC3339, expiresStr)
		if err != nil {
			return errResult(fmt.Errorf("invalid expires (want RFC3339): %w", err)), nil
		}
		token := &auth.AccessToken{
			Scope:   req.GetString("scope", "read write"),
			Expires: expires,
		}
		return runWrite(ctx, req, "POST", "api/access-token/", token, func() (any, error) {
			return d.Tandoor.Auth().CreateAccessToken(ctx, token)
		})
	}

	deleteToken := mcpgo.NewTool("auth_delete_token",
		mcpgo.WithDescription("Delete an API access token."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Token ID")),
		dryRunParam(),
	)
	deleteTokenHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/access-token/%d/", id), func() error {
			return d.Tandoor.Auth().DeleteAccessToken(ctx, id)
		})
	}

	return []toolDef{
		{tool: listTokens, handler: listTokensHandler},
		{tool: getToken, handler: getTokenHandler},
		{tool: createToken, handler: createTokenHandler, write: true},
		{tool: deleteToken, handler: deleteTokenHandler, write: true},
	}
}
