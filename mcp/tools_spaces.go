package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/space"
)

func registerSpaceTools(d *deps) []toolDef {
	// --- spaces ---
	spaceListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List spaces in the Tandoor instance. Supports text query, ordering, and pagination. Use all=true for the full set; jq projects fields to keep output small.")}
	spaceListOpts = append(spaceListOpts, baselineListParams()...)
	spaceList := mcpgo.NewTool("space_list", spaceListOpts...)
	spaceListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &space.ListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.Spaces().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[space.Space], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.Spaces().List(ctx, &o2)
		})
	}

	spaceGet := mcpgo.NewTool("space_get",
		mcpgo.WithDescription("Get a single space by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Space ID")),
		jqParam(),
	)
	spaceGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Spaces().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	spaceCurrent := mcpgo.NewTool("space_current",
		mcpgo.WithDescription("Get the space bound to the current access token (the default space for this token)."),
		jqParam(),
	)
	spaceCurrentHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.Spaces().Current(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- users ---
	userList := mcpgo.NewTool("user_list",
		mcpgo.WithDescription("List users. Optionally restrict to members of the given space IDs (space_ids). Returns the full set as a single array."),
		mcpgo.WithArray("space_ids", mcpgo.WithIntegerItems(), mcpgo.Description("Only return users in these spaces (filter_list)")),
		jqParam(),
	)
	userListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.Users().List(ctx, req.GetIntSlice("space_ids", nil))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	userGet := mcpgo.NewTool("user_get",
		mcpgo.WithDescription("Get a single user by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("User ID")),
		jqParam(),
	)
	userGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Users().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- groups ---
	groupList := mcpgo.NewTool("group_list",
		mcpgo.WithDescription("List permission groups. Returns the full set as a single array."),
		jqParam(),
	)
	groupListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		res, err := d.Tandoor.Groups().List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	groupGet := mcpgo.NewTool("group_get",
		mcpgo.WithDescription("Get a single permission group by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Group ID")),
		jqParam(),
	)
	groupGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Groups().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- households ---
	hhListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List households. Supports text query, ordering, and pagination. Use all=true for the full set; jq projects fields to keep output small.")}
	hhListOpts = append(hhListOpts, baselineListParams()...)
	hhList := mcpgo.NewTool("household_list", hhListOpts...)
	hhListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := baseListOptions(req, d)
		first, err := d.Tandoor.Households().List(ctx, &opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[space.Household], error) {
			o2 := opts
			o2.Page = page
			return d.Tandoor.Households().List(ctx, &o2)
		})
	}

	hhGet := mcpgo.NewTool("household_get",
		mcpgo.WithDescription("Get a single household by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Household ID")),
		jqParam(),
	)
	hhGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.Households().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- invite links ---
	invList := mcpgo.NewTool("invite_link_list",
		mcpgo.WithDescription("List space invite links (unused by default; set used=true to include used links)."),
		mcpgo.WithInteger("page", mcpgo.Description("Page number, 1-indexed (default 1)")),
		mcpgo.WithInteger("page_size", mcpgo.Description("Results per page (default 50)")),
		mcpgo.WithString("internal_note", mcpgo.Description("Filter by internal note (exact)")),
		mcpgo.WithBoolean("used", mcpgo.Description("true = include used links (omit for unused only)")),
		jqParam(),
	)
	invListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &space.InviteListOptions{
			Page:         req.GetInt("page", 1),
			PageSize:     req.GetInt("page_size", d.Cfg.DefaultPageSize),
			InternalNote: req.GetString("internal_note", ""),
		}
		if used, ok := boolArg(req, "used"); ok {
			opts.Used = &used
		}
		res, err := d.Tandoor.InviteLinks().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	invGet := mcpgo.NewTool("invite_link_get",
		mcpgo.WithDescription("Get a single invite link by ID (includes the invite URL)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Invite link ID")),
		jqParam(),
	)
	invGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.InviteLinks().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- user patch ---
	userPatch := mcpgo.NewTool("user_patch",
		mcpgo.WithDescription("Update a user's first/last name. Only these fields are writable via the API."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("User ID")),
		mcpgo.WithString("first_name", mcpgo.Description("First name")),
		mcpgo.WithString("last_name", mcpgo.Description("Last name")),
		dryRunParam(),
	)
	userPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		u := &space.User{ID: id}
		if v := req.GetString("first_name", ""); v != "" {
			u.FirstName = v
		}
		if v := req.GetString("last_name", ""); v != "" {
			u.LastName = v
		}
		body := map[string]any{"id": id}
		if u.FirstName != "" {
			body["first_name"] = u.FirstName
		}
		if u.LastName != "" {
			body["last_name"] = u.LastName
		}
		if len(body) == 1 {
			return errResult(fmt.Errorf("at least one of first_name or last_name is required")), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/user/%d/", id), body, func() (any, error) {
			return d.Tandoor.Users().Patch(ctx, u)
		})
	}

	// --- household writes ---
	hhCreate := mcpgo.NewTool("household_create",
		mcpgo.WithDescription("Create a household in the current space."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Household name")),
		dryRunParam(),
	)
	hhCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		h := &space.Household{Name: name}
		return runWrite(ctx, req, "POST", "api/household/", h, func() (any, error) {
			return d.Tandoor.Households().Create(ctx, h)
		})
	}

	hhUpdate := mcpgo.NewTool("household_update",
		mcpgo.WithDescription("Update a household's name (the only writable field). Returns the updated household."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Household ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("New name")),
		dryRunParam(),
	)
	hhUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		h := &space.Household{ID: id, Name: name}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/household/%d/", id), h, func() (any, error) {
			return d.Tandoor.Households().Update(ctx, h)
		})
	}

	hhPatch := mcpgo.NewTool("household_patch",
		mcpgo.WithDescription("Partially update a household's name (the only writable field)."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Household ID")),
		mcpgo.WithString("name", mcpgo.Description("New name")),
		dryRunParam(),
	)
	hhPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		h := &space.Household{ID: id, Name: req.GetString("name", "")}
		body := map[string]any{}
		if h.Name != "" {
			body["name"] = h.Name
		}
		if len(body) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/household/%d/", id), body, func() (any, error) {
			return d.Tandoor.Households().Patch(ctx, h)
		})
	}

	hhDelete := mcpgo.NewTool("household_delete",
		mcpgo.WithDescription("Delete a household."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Household ID")),
		dryRunParam(),
	)
	hhDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/household/%d/", id), func() error {
			return d.Tandoor.Households().Delete(ctx, id)
		})
	}

	// --- invite link writes ---
	invCreate := mcpgo.NewTool("invite_link_create",
		mcpgo.WithDescription("Create an invite link to join the space with a given group. Returns the invite URL."),
		mcpgo.WithInteger("group_id", mcpgo.Required(), mcpgo.Description("Group ID the invited user is assigned (see group_list)")),
		mcpgo.WithString("email", mcpgo.Description("Email to send the invite to (optional)")),
		mcpgo.WithInteger("household_id", mcpgo.Description("Assign the invited user to a household (optional)")),
		mcpgo.WithBoolean("reusable", mcpgo.Description("Allow the link to be used more than once (default false)")),
		mcpgo.WithString("internal_note", mcpgo.Description("Admin note on the link (optional)")),
		mcpgo.WithString("valid_until", mcpgo.Description("Expiry date YYYY-MM-DD (default: server default)")),
		dryRunParam(),
	)
	invCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		groupID, err := req.RequireInt("group_id")
		if err != nil {
			return errResult(err), nil
		}
		reqBody := &space.InviteLinkRequest{
			Email:        req.GetString("email", ""),
			GroupID:      groupID,
			Reusable:     req.GetBool("reusable", false),
			InternalNote: req.GetString("internal_note", ""),
		}
		if v := req.GetInt("household_id", 0); v != 0 {
			reqBody.HouseholdID = &v
		}
		if v := req.GetString("valid_until", ""); v != "" {
			reqBody.ValidUntil = &v
		}
		return runWrite(ctx, req, "POST", "api/invite-link/", reqBody, func() (any, error) {
			return d.Tandoor.InviteLinks().Create(ctx, reqBody)
		})
	}

	invDelete := mcpgo.NewTool("invite_link_delete",
		mcpgo.WithDescription("Delete an invite link."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Invite link ID")),
		dryRunParam(),
	)
	invDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/invite-link/%d/", id), func() error {
			return d.Tandoor.InviteLinks().Delete(ctx, id)
		})
	}

	return []toolDef{
		{tool: spaceList, handler: spaceListHandler},
		{tool: spaceGet, handler: spaceGetHandler},
		{tool: spaceCurrent, handler: spaceCurrentHandler},
		{tool: userList, handler: userListHandler},
		{tool: userGet, handler: userGetHandler},
		{tool: groupList, handler: groupListHandler},
		{tool: groupGet, handler: groupGetHandler},
		{tool: userPatch, handler: userPatchHandler, write: true},
		{tool: hhList, handler: hhListHandler},
		{tool: hhGet, handler: hhGetHandler},
		{tool: hhCreate, handler: hhCreateHandler, write: true},
		{tool: hhUpdate, handler: hhUpdateHandler, write: true},
		{tool: hhPatch, handler: hhPatchHandler, write: true},
		{tool: hhDelete, handler: hhDeleteHandler, write: true},
		{tool: invList, handler: invListHandler},
		{tool: invGet, handler: invGetHandler},
		{tool: invCreate, handler: invCreateHandler, write: true},
		{tool: invDelete, handler: invDeleteHandler, write: true},
	}
}

// baseListOptions builds the pagination.ListOptions baseline from the
// standard list parameters (query, page, page_size, ordering).
func baseListOptions(req mcpgo.CallToolRequest, d *deps) pagination.ListOptions {
	return pagination.ListOptions{
		Page:     req.GetInt("page", 1),
		PageSize: req.GetInt("page_size", d.Cfg.DefaultPageSize),
		Search:   req.GetString("query", ""),
		OrderBy:  req.GetString("ordering", ""),
	}
}

// baselineListParams are the shared list-tool parameters (query, page,
// page_size, ordering, all, jq).
func baselineListParams() []mcpgo.ToolOption {
	return []mcpgo.ToolOption{
		mcpgo.WithString("query", mcpgo.Description("Text search (icontains/trigram name lookup)")),
		mcpgo.WithInteger("page", mcpgo.Description("Page number, 1-indexed (default 1)")),
		mcpgo.WithInteger("page_size", mcpgo.Description("Results per page (default 50)")),
		mcpgo.WithString("ordering", mcpgo.Description("Ordering field, e.g. name, -name")),
		mcpgo.WithBoolean("all", mcpgo.Description("Auto-paginate and return a flat array (overrides page/page_size)")),
		mcpgo.WithString("jq", mcpgo.Description("gojq filter applied to the result before returning")),
	}
}
