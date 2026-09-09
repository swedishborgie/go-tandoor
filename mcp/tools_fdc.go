package mcp

import (
	"context"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/fdc"
)

// registerFdcTools registers the USDA FoodData Central tools. Both require
// FDC_API_KEY at server construction; without it they return a clear error.
func registerFdcTools(d *deps) []toolDef {
	search := mcpgo.NewTool("fdc_search",
		mcpgo.WithDescription("Search the USDA FoodData Central database by name. Returns FDC IDs and abridged nutrient data; pair with fdc_get_food for full detail."),
		mcpgo.WithString("query", mcpgo.Required(), mcpgo.Description("Search query, e.g. a food name")),
		mcpgo.WithInteger("limit", mcpgo.Description("Maximum results (default 10, FDC max 30)")),
		mcpgo.WithArray("types", mcpgo.WithStringItems(), mcpgo.Description("Optional data type filter: srlegacy, foundation, survey, branded")),
		jqParam(),
	)
	searchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if d.FDC == nil {
			return errResult(fdcNotConfigured()), nil
		}
		query := req.GetString("query", "")
		if query == "" {
			return errResult(fmt.Errorf("query is required")), nil
		}
		limit := req.GetInt("limit", 10)
		if limit > 30 {
			limit = 30
		}
		dataTypes, err := parseFdcDataTypes(req.GetStringSlice("types", nil))
		if err != nil {
			return errResult(err), nil
		}
		resp, err := d.FDC.SearchFoods(ctx, query, dataTypes, &limit, nil, "", "", "")
		if err != nil {
			return errResult(err), nil
		}
		if jq := req.GetString("jq", ""); jq != "" {
			s, err := applyJQ(ctx, jq, resp)
			if err != nil {
				return errResult(err), nil
			}
			return mcpgo.NewToolResultText(s), nil
		}
		return jsonResult(resp), nil
	}

	getFood := mcpgo.NewTool("fdc_get_food",
		mcpgo.WithDescription("Get a single FDC food with full nutrient detail by FDC ID (from fdc_search)."),
		mcpgo.WithInteger("fdc_id", mcpgo.Required(), mcpgo.Description("FDC food ID")),
		mcpgo.WithArray("nutrients", mcpgo.WithIntegerItems(), mcpgo.Description("Optional nutrient IDs to include; defaults to all")),
		jqParam(),
	)
	getFoodHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if d.FDC == nil {
			return errResult(fdcNotConfigured()), nil
		}
		fdcID := req.GetInt("fdc_id", 0)
		if fdcID == 0 {
			return errResult(fmt.Errorf("fdc_id is required")), nil
		}
		nutrients := req.GetIntSlice("nutrients", nil)
		food, err := d.FDC.GetFood(ctx, fdcID, "", nutrients)
		if err != nil {
			return errResult(err), nil
		}
		if jq := req.GetString("jq", ""); jq != "" {
			s, err := applyJQ(ctx, jq, food)
			if err != nil {
				return errResult(err), nil
			}
			return mcpgo.NewToolResultText(s), nil
		}
		return jsonResult(food), nil
	}

	return []toolDef{
		{tool: search, handler: searchHandler},
		{tool: getFood, handler: getFoodHandler},
	}
}

// parseFdcDataTypes maps the CLI-style type aliases (srlegacy, foundation,
// survey, branded) to fdc.DataType values.
func parseFdcDataTypes(names []string) ([]fdc.DataType, error) {
	if len(names) == 0 {
		return nil, nil
	}
	aliases := map[string]fdc.DataType{
		"srlegacy":   fdc.DataTypeSRLegacy,
		"sr legacy":  fdc.DataTypeSRLegacy,
		"foundation": fdc.DataTypeFoundation,
		"survey":     fdc.DataTypeSurvey,
		"branded":    fdc.DataTypeBranded,
	}
	out := make([]fdc.DataType, 0, len(names))
	for _, n := range names {
		key := strings.ToLower(strings.TrimSpace(n))
		dt, ok := aliases[key]
		if !ok {
			return nil, fmt.Errorf("unknown data type %q (valid: srlegacy, foundation, survey, branded)", n)
		}
		out = append(out, dt)
	}
	return out, nil
}
