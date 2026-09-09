package auditfood

import (
	"context"
	"fmt"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/idref"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/swedishborgie/go-tandoor/unit"
)

// DefaultPer100Amount is the default property basis amount (per 100 g).
const DefaultPer100Amount = 100.0

// AttachOptions configures Attach.
type AttachOptions struct {
	// FoodID or IngredientID selects the target (exactly one). An
	// ingredient resolves to the food it references.
	FoodID       int
	IngredientID int

	PropertyTypeID int
	Amount         float64

	// Per100Amount sets the food's properties_food_amount (default 100).
	Per100Amount float64
	// Per100UnitID sets the food's properties_food_unit. When 0, the food's
	// existing unit is kept if present, otherwise the unit named "gram" is
	// resolved from the instance (error when missing).
	Per100UnitID int
}

// Per100Basis is the per-amount property basis applied to the food.
type Per100Basis struct {
	Amount float64 `json:"amount"`
	UnitID int     `json:"unit_id"`
}

// AttachResult reports the attach outcome.
type AttachResult struct {
	Action       string             `json:"action"` // "created" or "updated"
	Property     *property.Property `json:"property,omitempty"`
	Food         *food.Food         `json:"food,omitempty"`
	Per100       *Per100Basis       `json:"per100,omitempty"`
	ActionsTaken []string           `json:"actions_taken"`
}

// ResolveTargetFoodID returns the food ID for a food/ingredient target.
func ResolveTargetFoodID(ctx context.Context, c *tandoor.Client, foodID, ingredientID int) (int, error) {
	switch {
	case foodID != 0 && ingredientID != 0:
		return 0, fmt.Errorf("set food_id or ingredient_id, not both")
	case foodID != 0:
		return foodID, nil
	case ingredientID != 0:
		ing, err := c.Ingredients().Get(ctx, ingredientID)
		if err != nil {
			return 0, fmt.Errorf("get ingredient %d: %w", ingredientID, err)
		}
		if ing.Food == nil {
			return 0, fmt.Errorf("ingredient %d has no food attached", ingredientID)
		}
		return ing.Food.ID, nil
	default:
		return 0, fmt.Errorf("set food_id or ingredient_id")
	}
}

// Attach attaches (or updates) a property value on a food, idempotently.
// The property type is fetched authoritatively from the server — Tandoor's
// writable nested serializer overwrites every field sent, so a fabricated
// type would corrupt it. When dryRun, the planned action is returned without
// writing.
func Attach(ctx context.Context, c *tandoor.Client, opts *AttachOptions, dryRun bool) (*AttachResult, error) {
	if opts.PropertyTypeID == 0 {
		return nil, fmt.Errorf("property_type_id is required")
	}
	foodID, err := ResolveTargetFoodID(ctx, c, opts.FoodID, opts.IngredientID)
	if err != nil {
		return nil, err
	}

	propType, err := c.PropertyTypes().Get(ctx, opts.PropertyTypeID)
	if err != nil {
		return nil, fmt.Errorf("get property type %d: %w (create it first with property_type_create)", opts.PropertyTypeID, err)
	}
	f, err := c.Foods().Get(ctx, foodID)
	if err != nil {
		return nil, fmt.Errorf("get food %d: %w", foodID, err)
	}

	res := &AttachResult{ActionsTaken: []string{}}

	// The food endpoint returns the authoritative properties array.
	var existing *property.Property
	for i := range f.Properties {
		if f.Properties[i].Type.ID == opts.PropertyTypeID {
			existing = &f.Properties[i]
			break
		}
	}

	if dryRun {
		if existing != nil {
			res.Action = "would_update"
		} else {
			res.Action = "would_create"
		}
		return res, nil
	}

	amount := opts.Amount
	if existing != nil {
		// Targeted property PATCH (Tandoor requires property_type on PATCH).
		updated, err := c.Properties().Patch(ctx, &property.Property{
			ID:             existing.ID,
			Type:           *propType,
			PropertyAmount: &amount,
		})
		if err != nil {
			return nil, fmt.Errorf("patch property %d: %w", existing.ID, err)
		}
		res.Action = "updated"
		res.Property = updated
		res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("property_%d_updated", existing.ID))
	} else {
		// POST /api/property/ ignores the food field, so the new property is
		// added through the food's properties array (full-array PATCH).
		f.Properties = append(f.Properties, property.Property{
			Type:           *propType,
			PropertyAmount: &amount,
		})
		basis, err := applyPer100(ctx, c, f, opts.Per100Amount, opts.Per100UnitID)
		if err != nil {
			return nil, err
		}
		res.Per100 = basis
		updatedFood, err := c.Foods().Patch(ctx, f)
		if err != nil {
			return nil, fmt.Errorf("patch food %d with properties: %w", foodID, err)
		}
		res.Action = "created"
		res.Food = updatedFood
		for i := range updatedFood.Properties {
			if updatedFood.Properties[i].Type.ID == opts.PropertyTypeID {
				res.Property = &updatedFood.Properties[i]
				break
			}
		}
		res.ActionsTaken = append(res.ActionsTaken, "food_properties_patched")
		return res, nil
	}

	// Update path: keep the per-100 basis in sync (skip when unchanged).
	basis, err := applyPer100(ctx, c, f, opts.Per100Amount, opts.Per100UnitID)
	if err != nil {
		res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("warn: per-100 basis unchanged: %v", err))
	} else if f.PropertiesFoodAmount == nil || *f.PropertiesFoodAmount != basis.Amount || unitIDFromRef(f.PropertiesFoodUnit) != basis.UnitID {
		if _, err := c.Foods().Patch(ctx, &food.Food{
			ID:                   f.ID,
			PropertiesFoodAmount: &basis.Amount,
			PropertiesFoodUnit:   basis.UnitID,
		}); err != nil {
			res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("warn: could not set per-100 basis: %v", err))
		} else {
			res.Per100 = basis
			res.ActionsTaken = append(res.ActionsTaken, "per100_basis_updated")
		}
	}
	res.Food = f
	return res, nil
}

// applyPer100 resolves the per-100 basis (keeping the food's existing unit
// unless overridden) and sets the fields on f. It errors when no unit can be
// resolved: the caller must supply Per100UnitID on instances without a unit
// named "gram".
func applyPer100(ctx context.Context, c *tandoor.Client, f *food.Food, amount float64, unitID int) (*Per100Basis, error) {
	if amount <= 0 {
		amount = DefaultPer100Amount
	}
	if unitID == 0 {
		if id := unitIDFromRef(f.PropertiesFoodUnit); id != 0 {
			unitID = id
		} else if id, err := findUnitIDByName(ctx, c, "gram"); err == nil {
			unitID = id
		} else {
			return nil, fmt.Errorf("cannot resolve per-100 unit: food has no properties_food_unit and no unit named %q found (%w) — set per_100_unit_id", "gram", err)
		}
	}
	f.PropertiesFoodAmount = &amount
	f.PropertiesFoodUnit = unitID
	return &Per100Basis{Amount: amount, UnitID: unitID}, nil
}

// unitIDFromRef extracts a unit ID from a properties_food_unit value (bare
// int or nested object, as returned by the API).
func unitIDFromRef(v any) int {
	return idref.AsAnyID(v)
}

// FindUnitIDByName returns the ID of the unit whose name matches (case
// insensitive exact match), auto-paginating the unit list.
func FindUnitIDByName(ctx context.Context, c *tandoor.Client, name string) (int, error) {
	return findUnitIDByName(ctx, c, name)
}

func findUnitIDByName(ctx context.Context, c *tandoor.Client, name string) (int, error) {
	svc := c.Units()
	page, err := svc.List(ctx, &unit.ListOptions{ListOptions: pagination.ListOptions{PageSize: 100}})
	if err != nil {
		return 0, fmt.Errorf("list units: %w", err)
	}
	all, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[unit.Unit], error) {
		return svc.List(ctx, &unit.ListOptions{ListOptions: pagination.ListOptions{Page: pageNum, PageSize: 100}})
	})
	if err != nil {
		return 0, err
	}
	for _, u := range all {
		if strings.EqualFold(u.Name, name) {
			return u.ID, nil
		}
	}
	return 0, fmt.Errorf("no unit named %q", name)
}
