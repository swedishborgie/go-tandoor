package auditfood

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/swedishborgie/go-tandoor/property"
)

// FdcAttachOptions configures AttachFDCProperties.
type FdcAttachOptions struct {
	FoodID int
	// FDCID overrides the food's FDC ID (e.g. from a food_inspect match).
	FDCID *int
	// Per100Amount / Per100UnitID behave as in AttachOptions.
	Per100Amount float64
	Per100UnitID int
}

// FdcAttachedProperty is one property attached from the FDC lookup.
type FdcAttachedProperty struct {
	PropertyTypeID int     `json:"property_type_id"`
	Name           string  `json:"name"`
	FDCID          uint    `json:"fdc_id"`
	Amount         float64 `json:"amount"`
	Action         string  `json:"action"` // "created" or "updated"
}

// FdcAttachResult reports the bulk FDC attach outcome.
type FdcAttachResult struct {
	FoodID       int                   `json:"food_id"`
	FDCID        int                   `json:"fdc_id"`
	Attached     []FdcAttachedProperty `json:"attached"`
	Skipped      []string              `json:"skipped,omitempty"`
	Per100       *Per100Basis          `json:"per100,omitempty"`
	ActionsTaken []string              `json:"actions_taken"`
}

// AttachFDCProperties fetches a food's FDC record and attaches every
// property whose type has a matching FDC ID, in a single food PATCH.
// Property types are enumerated from the instance (auto-paginated).
// dryRun returns the plan without writing.
func AttachFDCProperties(ctx context.Context, c *tandoor.Client, fdcClient *fdc.Client, opts *FdcAttachOptions, dryRun bool) (*FdcAttachResult, error) {
	if fdcClient == nil {
		return nil, fmt.Errorf("FDC API key not set")
	}
	f, err := c.Foods().Get(ctx, opts.FoodID)
	if err != nil {
		return nil, fmt.Errorf("get food: %w", err)
	}
	fdcID := 0
	if opts.FDCID != nil {
		fdcID = *opts.FDCID
	} else if f.FDCID != nil {
		fdcID = *f.FDCID
	}
	if fdcID == 0 {
		return nil, fmt.Errorf("food %d has no FDC ID and none provided", opts.FoodID)
	}

	fdcFood, err := fdcClient.GetFood(ctx, fdcID, "", nil)
	if err != nil {
		return nil, fmt.Errorf("get FDC food %d: %w", fdcID, err)
	}

	types, err := FetchAllPropertyTypes(ctx, c)
	if err != nil {
		return nil, err
	}

	// Map each FDC nutrient to the property type that carries it.
	fdcByID := make(map[uint]fdc.FoodNutrient, len(fdcFood.FoodNutrients))
	for _, n := range fdcFood.FoodNutrients {
		if n.Amount != nil && n.Nutrient != nil {
			fdcByID[n.ID] = n
		}
	}

	res := &FdcAttachResult{FoodID: opts.FoodID, FDCID: fdcID, Attached: []FdcAttachedProperty{}, ActionsTaken: []string{}}
	changed := false
	for _, t := range types {
		if t.FDCID == nil {
			continue
		}
		n, ok := fdcByID[uint(*t.FDCID)]
		if !ok || n.Amount == nil {
			continue
		}
		amount := *n.Amount
		// Update in place if the food already has this property type, else append.
		var existing *property.Property
		for i := range f.Properties {
			if f.Properties[i].Type.ID == t.ID {
				existing = &f.Properties[i]
				break
			}
		}
		if existing != nil {
			existing.PropertyAmount = &amount
		} else {
			f.Properties = append(f.Properties, property.Property{Type: t, PropertyAmount: &amount})
		}
		action := "created"
		if existing != nil {
			action = "updated"
		}
		res.Attached = append(res.Attached, FdcAttachedProperty{
			PropertyTypeID: t.ID,
			Name:           t.Name,
			FDCID:          n.ID,
			Amount:         amount,
			Action:         action,
		})
		changed = true
	}
	if !changed {
		res.Skipped = []string{"no FDC nutrients matched property types with fdc_id"}
		return res, nil
	}

	if dryRun {
		res.ActionsTaken = append(res.ActionsTaken, "would_patch_food_properties")
		return res, nil
	}

	// Single food PATCH with the full properties array + per-100 basis.
	basis, err := applyPer100(ctx, c, f, opts.Per100Amount, opts.Per100UnitID)
	if err != nil {
		return res, err
	}
	res.Per100 = basis
	if _, err := c.Foods().Patch(ctx, f); err != nil {
		return res, fmt.Errorf("patch food %d: %w", opts.FoodID, err)
	}
	res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("food_properties_patched(%d)", len(f.Properties)))
	return res, nil
}
