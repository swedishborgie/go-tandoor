package auditfood

import (
	"context"
	"fmt"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/property"
)

// ManyProperty is one property value for AttachMany. Exactly one of
// PropertyTypeID / PropertyTypeName must be set; names are resolved
// case-insensitively against the instance's property types.
type ManyProperty struct {
	PropertyTypeID   int
	PropertyTypeName string
	Amount           float64
}

// AttachManyOptions configures AttachMany.
type AttachManyOptions struct {
	FoodID int
	// Properties is the set of per-100-g values to attach. Duplicate
	// property types (by id or name) collapse, with the last entry winning.
	Properties   []ManyProperty
	Per100Amount float64
	Per100UnitID int
}

// AttachedProperty is one property applied by AttachMany.
type AttachedProperty struct {
	PropertyTypeID int     `json:"property_type_id"`
	Name           string  `json:"name"`
	Amount         float64 `json:"amount"`
	Action         string  `json:"action"` // "created" or "updated"
}

// AttachManyResult reports the batch attach outcome.
type AttachManyResult struct {
	FoodID       int                `json:"food_id"`
	Attached     []AttachedProperty `json:"attached"`
	Per100       *Per100Basis       `json:"per100,omitempty"`
	ActionsTaken []string           `json:"actions_taken"`
}

// AttachMany attaches (or updates) several per-100-g properties on one food
// in a single food PATCH — the batch counterpart of Attach, for when FDC
// lacks nutrients and the values must be attached manually. Property types
// may be given by ID or by name. Idempotent: re-running updates the same
// properties.
func AttachMany(ctx context.Context, c *tandoor.Client, opts *AttachManyOptions) (*AttachManyResult, error) {
	if opts.FoodID == 0 {
		return nil, fmt.Errorf("food_id is required")
	}
	if len(opts.Properties) == 0 {
		return nil, fmt.Errorf("at least one property is required")
	}

	types, err := FetchAllPropertyTypes(ctx, c)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]property.Type, len(types))
	byName := make(map[string]property.Type, len(types))
	for _, t := range types {
		byID[t.ID] = t
		byName[strings.ToLower(t.Name)] = t
	}

	// Resolve and collapse (last wins, preserving first-seen order).
	type entry struct {
		pt     property.Type
		amount float64
	}
	order := make([]int, 0, len(opts.Properties))
	m := make(map[int]*entry, len(opts.Properties))
	for i, p := range opts.Properties {
		var pt *property.Type
		switch {
		case p.PropertyTypeID != 0:
			t, ok := byID[p.PropertyTypeID]
			if !ok {
				return nil, fmt.Errorf("properties[%d]: unknown property_type_id %d (see property_type_list)", i, p.PropertyTypeID)
			}
			pt = &t
		case p.PropertyTypeName != "":
			t, ok := byName[strings.ToLower(p.PropertyTypeName)]
			if !ok {
				return nil, fmt.Errorf("properties[%d]: no property type named %q (see property_type_list)", i, p.PropertyTypeName)
			}
			pt = &t
		default:
			return nil, fmt.Errorf("properties[%d]: set property_type_id or property_type_name", i)
		}
		if e, ok := m[pt.ID]; ok {
			e.amount = p.Amount
		} else {
			m[pt.ID] = &entry{pt: *pt, amount: p.Amount}
			order = append(order, pt.ID)
		}
	}

	f, err := c.Foods().Get(ctx, opts.FoodID)
	if err != nil {
		return nil, fmt.Errorf("get food %d: %w", opts.FoodID, err)
	}

	res := &AttachManyResult{FoodID: opts.FoodID, Attached: []AttachedProperty{}, ActionsTaken: []string{}}
	for _, id := range order {
		e := m[id]
		var existing *property.Property
		for i := range f.Properties {
			if f.Properties[i].Type.ID == e.pt.ID {
				existing = &f.Properties[i]
				break
			}
		}
		if existing != nil {
			existing.PropertyAmount = &e.amount
			res.Attached = append(res.Attached, AttachedProperty{PropertyTypeID: e.pt.ID, Name: e.pt.Name, Amount: e.amount, Action: "updated"})
		} else {
			f.Properties = append(f.Properties, property.Property{Type: e.pt, PropertyAmount: &e.amount})
			res.Attached = append(res.Attached, AttachedProperty{PropertyTypeID: e.pt.ID, Name: e.pt.Name, Amount: e.amount, Action: "created"})
		}
	}

	basis, err := applyPer100(ctx, c, f, opts.Per100Amount, opts.Per100UnitID)
	if err != nil {
		return res, err
	}
	res.Per100 = basis
	if _, err := c.Foods().Patch(ctx, f); err != nil {
		return res, fmt.Errorf("patch food %d: %w", opts.FoodID, err)
	}
	res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("food_properties_patched(%d)", len(res.Attached)))
	return res, nil
}
