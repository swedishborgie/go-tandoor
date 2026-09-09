package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/swedishborgie/go-tandoor/inventory"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func registerInventoryTools(d *deps) []toolDef {
	// --- locations ---
	locListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List inventory locations (fridge, pantry, ...). Use all=true for the full set; jq projects fields to keep output small.")}
	locListOpts = append(locListOpts, baselineListParams()...)
	locList := mcpgo.NewTool("inventory_location_list", locListOpts...)
	locListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		opts := &inventory.LocationListOptions{ListOptions: baseListOptions(req, d)}
		first, err := d.Tandoor.InventoryLocations().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[inventory.Location], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.InventoryLocations().List(ctx, &o2)
		})
	}

	locGet := mcpgo.NewTool("inventory_location_get",
		mcpgo.WithDescription("Get a single inventory location by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Location ID")),
		jqParam(),
	)
	locGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.InventoryLocations().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- entries ---
	entryListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List inventory entries (food quantities on hand). Filter by food, location, or barcode; include_empty controls zero-quantity rows. Use all=true for the full set; jq projects fields to keep output small.")}
	entryListOpts = append(entryListOpts, baselineListParams()...)
	entryListOpts = append(entryListOpts,
		mcpgo.WithInteger("food_id", mcpgo.Description("Filter by food ID")),
		mcpgo.WithInteger("location_id", mcpgo.Description("Filter by inventory location ID")),
		mcpgo.WithString("code", mcpgo.Description("Filter by barcode")),
		mcpgo.WithBoolean("include_empty", mcpgo.Description("Include zero-quantity entries")),
	)
	entryList := mcpgo.NewTool("inventory_entry_list", entryListOpts...)
	entryListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		opts := &inventory.EntryListOptions{ListOptions: &base}
		if id := req.GetInt("food_id", 0); id != 0 {
			opts.FoodID = &id
		}
		if id := req.GetInt("location_id", 0); id != 0 {
			opts.LocationID = &id
		}
		if s := req.GetString("code", ""); s != "" {
			opts.Code = &s
		}
		if v, ok := boolArg(req, "include_empty"); ok {
			opts.IncludeEmpty = &v
		}
		first, err := d.Tandoor.InventoryEntries().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[inventory.Entry], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.InventoryEntries().List(ctx, &o2)
		})
	}

	entryGet := mcpgo.NewTool("inventory_entry_get",
		mcpgo.WithDescription("Get a single inventory entry by ID."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Inventory entry ID")),
		jqParam(),
	)
	entryGetHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		res, err := d.Tandoor.InventoryEntries().Get(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResultJQ(ctx, req, res)
	}

	// --- logs ---
	logListOpts := []mcpgo.ToolOption{mcpgo.WithDescription("List inventory logs (stock changes). Filter by entry or food. Use all=true for the full set; jq projects fields to keep output small.")}
	logListOpts = append(logListOpts, baselineListParams()...)
	logListOpts = append(logListOpts,
		mcpgo.WithInteger("entry_id", mcpgo.Description("Filter by inventory entry ID")),
		mcpgo.WithInteger("food_id", mcpgo.Description("Filter by food ID")),
	)
	logList := mcpgo.NewTool("inventory_log_list", logListOpts...)
	logListHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		base := baseListOptions(req, d)
		opts := &inventory.LogListOptions{ListOptions: &base}
		if id := req.GetInt("entry_id", 0); id != 0 {
			opts.EntryID = &id
		}
		if id := req.GetInt("food_id", 0); id != 0 {
			opts.FoodID = &id
		}
		first, err := d.Tandoor.InventoryLogs().List(ctx, opts)
		if err != nil {
			return errResult(err), nil
		}
		return listAllResult(ctx, req, first, func(page int) (*pagination.Paginated[inventory.Log], error) {
			o2 := *opts
			o2.Page = page
			return d.Tandoor.InventoryLogs().List(ctx, &o2)
		})
	}

	// --- location writes ---
	locCreate := mcpgo.NewTool("inventory_location_create",
		mcpgo.WithDescription("Create an inventory location (e.g. fridge, pantry). Returns the created location."),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Location name")),
		mcpgo.WithBoolean("is_freezer", mcpgo.Description("Whether this is a freezer location")),
		mcpgo.WithInteger("household_id", mcpgo.Required(), mcpgo.Description("Household ID (see household_list)")),
		dryRunParam(),
	)
	locCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		householdID, err := req.RequireInt("household_id")
		if err != nil {
			return errResult(err), nil
		}
		loc := &inventory.Location{Name: name, Household: intPtr(householdID)}
		if v, err := req.RequireBool("is_freezer"); err == nil {
			loc.IsFreezer = v
		}
		return runWrite(ctx, req, "POST", "api/inventory-location/", loc, func() (any, error) {
			return d.Tandoor.InventoryLocations().Create(ctx, loc)
		})
	}

	locUpdate := mcpgo.NewTool("inventory_location_update",
		mcpgo.WithDescription("Update an inventory location (full replacement). Returns the updated location."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Location ID")),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Location name")),
		mcpgo.WithBoolean("is_freezer", mcpgo.Description("Whether this is a freezer location")),
		mcpgo.WithInteger("household_id", mcpgo.Required(), mcpgo.Description("Household ID")),
		dryRunParam(),
	)
	locUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return errResult(err), nil
		}
		householdID, err := req.RequireInt("household_id")
		if err != nil {
			return errResult(err), nil
		}
		loc := &inventory.Location{ID: id, Name: name, Household: intPtr(householdID)}
		if v, err := req.RequireBool("is_freezer"); err == nil {
			loc.IsFreezer = v
		}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/inventory-location/%d/", id), loc, func() (any, error) {
			return d.Tandoor.InventoryLocations().Update(ctx, loc)
		})
	}

	locPatch := mcpgo.NewTool("inventory_location_patch",
		mcpgo.WithDescription("Partially update an inventory location. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Location ID")),
		mcpgo.WithString("name", mcpgo.Description("Location name")),
		mcpgo.WithBoolean("is_freezer", mcpgo.Description("Whether this is a freezer location")),
		mcpgo.WithInteger("household_id", mcpgo.Description("Household ID")),
		dryRunParam(),
	)
	locPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		loc := &inventory.Location{ID: id}
		if v, err := req.RequireString("name"); err == nil {
			loc.Name = v
		}
		if v, err := req.RequireBool("is_freezer"); err == nil {
			loc.IsFreezer = v
		}
		if v, err := req.RequireInt("household_id"); err == nil {
			loc.Household = &v
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/inventory-location/%d/", id), loc, func() (any, error) {
			return d.Tandoor.InventoryLocations().Patch(ctx, loc)
		})
	}

	locDelete := mcpgo.NewTool("inventory_location_delete",
		mcpgo.WithDescription("Delete an inventory location. Fails if entries still reference it."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Location ID")),
		dryRunParam(),
	)
	locDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/inventory-location/%d/", id), func() error {
			return d.Tandoor.InventoryLocations().Delete(ctx, id)
		})
	}

	// --- entry writes ---
	entryWriteFields := []mcpgo.ToolOption{
		mcpgo.WithInteger("food_id", mcpgo.Description("Food ID")),
		mcpgo.WithInteger("unit_id", mcpgo.Description("Unit ID (required for create/update — the API rejects entries without a unit)")),
		mcpgo.WithNumber("amount", mcpgo.Description("Quantity on hand")),
		mcpgo.WithString("note", mcpgo.Description("Note")),
		mcpgo.WithString("sub_location", mcpgo.Description("Sub-location (e.g. door shelf)")),
		mcpgo.WithString("expires", mcpgo.Description("Expiry date (RFC3339 or YYYY-MM-DD)")),
	}
	buildInventoryEntry := func(id int, req mcpgo.CallToolRequest) (*inventory.Entry, error) {
		e := &inventory.Entry{ID: id}
		if v, err := req.RequireInt("food_id"); err == nil && v != 0 {
			e.Food = &v
		}
		if v, err := req.RequireInt("unit_id"); err == nil && v != 0 {
			e.Unit = &v
		}
		if v, err := req.RequireFloat("amount"); err == nil {
			e.Amount = v
		}
		if v, err := req.RequireString("note"); err == nil {
			e.Note = v
		}
		if v, err := req.RequireString("sub_location"); err == nil {
			e.SubLocation = v
		}
		if v, err := req.RequireString("expires"); err == nil {
			t, err := parseDateOrDateTime(v)
			if err != nil {
				return nil, err
			}
			e.Expires = &t
		}
		return e, nil
	}

	einCreateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Create an inventory entry (item on hand). Returns the created entry."),
		mcpgo.WithInteger("food_id", mcpgo.Required(), mcpgo.Description("Food ID")),
		mcpgo.WithInteger("location_id", mcpgo.Required(), mcpgo.Description("Inventory location ID")),
		dryRunParam(),
	}, entryWriteFields...)
	einCreate := mcpgo.NewTool("inventory_entry_create", einCreateOpts...)
	einCreateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		locationID, err := req.RequireInt("location_id")
		if err != nil {
			return errResult(err), nil
		}
		e, err := buildInventoryEntry(0, req)
		if err != nil {
			return errResult(err), nil
		}
		e.Location = &inventory.Location{ID: locationID}
		return runWrite(ctx, req, "POST", "api/inventory-entry/", e, func() (any, error) {
			return d.Tandoor.InventoryEntries().Create(ctx, e)
		})
	}

	einUpdateOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Update an inventory entry (full replacement). Returns the updated entry."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Entry ID")),
		mcpgo.WithInteger("food_id", mcpgo.Required(), mcpgo.Description("Food ID")),
		mcpgo.WithInteger("location_id", mcpgo.Required(), mcpgo.Description("Inventory location ID")),
		dryRunParam(),
	}, entryWriteFields...)
	einUpdate := mcpgo.NewTool("inventory_entry_update", einUpdateOpts...)
	einUpdateHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		locationID, err := req.RequireInt("location_id")
		if err != nil {
			return errResult(err), nil
		}
		e, err := buildInventoryEntry(id, req)
		if err != nil {
			return errResult(err), nil
		}
		e.Location = &inventory.Location{ID: locationID}
		return runWrite(ctx, req, "PUT", fmt.Sprintf("api/inventory-entry/%d/", id), e, func() (any, error) {
			return d.Tandoor.InventoryEntries().Update(ctx, e)
		})
	}

	einPatchOpts := append([]mcpgo.ToolOption{
		mcpgo.WithDescription("Partially update an inventory entry. Only provided fields change."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Entry ID")),
		dryRunParam(),
	}, entryWriteFields...)
	einPatch := mcpgo.NewTool("inventory_entry_patch", einPatchOpts...)
	einPatchHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		e, err := buildInventoryEntry(id, req)
		if err != nil {
			return errResult(err), nil
		}
		if locationID, err := req.RequireInt("location_id"); err == nil {
			e.Location = &inventory.Location{ID: locationID}
		}
		return runWrite(ctx, req, "PATCH", fmt.Sprintf("api/inventory-entry/%d/", id), e, func() (any, error) {
			return d.Tandoor.InventoryEntries().Patch(ctx, e)
		})
	}

	einDelete := mcpgo.NewTool("inventory_entry_delete",
		mcpgo.WithDescription("Delete an inventory entry."),
		mcpgo.WithInteger("id", mcpgo.Required(), mcpgo.Description("Entry ID")),
		dryRunParam(),
	)
	einDeleteHandler := func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		id, err := req.RequireInt("id")
		if err != nil {
			return errResult(err), nil
		}
		return deleteResult(ctx, req, fmt.Sprintf("api/inventory-entry/%d/", id), func() error {
			return d.Tandoor.InventoryEntries().Delete(ctx, id)
		})
	}

	return []toolDef{
		{tool: locList, handler: locListHandler},
		{tool: locGet, handler: locGetHandler},
		{tool: entryList, handler: entryListHandler},
		{tool: entryGet, handler: entryGetHandler},
		{tool: logList, handler: logListHandler},
		{tool: locCreate, handler: locCreateHandler, write: true},
		{tool: locUpdate, handler: locUpdateHandler, write: true},
		{tool: locPatch, handler: locPatchHandler, write: true},
		{tool: locDelete, handler: locDeleteHandler, write: true},
		{tool: einCreate, handler: einCreateHandler, write: true},
		{tool: einUpdate, handler: einUpdateHandler, write: true},
		{tool: einPatch, handler: einPatchHandler, write: true},
		{tool: einDelete, handler: einDeleteHandler, write: true},
	}
}
