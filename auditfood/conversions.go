package auditfood

import (
	"context"
	"fmt"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/unit"
)

// AutoConversionsResult reports the unit conversions created (or planned).
type AutoConversionsResult struct {
	FoodID       int               `json:"food_id"`
	BaseUnit     int               `json:"base_unit"`
	Created      []unit.Conversion `json:"created,omitempty"`
	Skipped      []string          `json:"skipped,omitempty"`
	ActionsTaken []string          `json:"actions_taken"`
}

// AutoConversions creates sensible unit conversions for a food based on its
// property unit (or the unit named "gram" when the food has no
// properties_food_unit):
//
//	gram  → 28.3495 g = 1 oz, 453.592 g = 1 lb, 1000 g = 1 kg
//	millilitre → 236.588 ml = 1 cup, 14.787 ml = 1 tbsp, 4.929 ml = 1 tsp
//
// Existing conversions are skipped; target units missing on the instance are
// skipped with a note. Unit IDs are resolved by name from the instance
// rather than hardcoded. dryRun returns the plan without writing.
func AutoConversions(ctx context.Context, c *tandoor.Client, foodID int, dryRun bool) (*AutoConversionsResult, error) {
	f, err := c.Foods().Get(ctx, foodID)
	if err != nil {
		return nil, fmt.Errorf("get food: %w", err)
	}

	// Base unit: the food's property unit, else the unit named "gram".
	baseUnit := unitIDFromRef(f.PropertiesFoodUnit)
	if baseUnit == 0 {
		baseUnit, err = FindUnitIDByName(ctx, c, "gram")
		if err != nil {
			return nil, fmt.Errorf("resolve base unit: %w", err)
		}
	}

	// All units on the instance, keyed by lowercased name.
	svc := c.Units()
	page, err := svc.List(ctx, &unit.ListOptions{ListOptions: pagination.ListOptions{PageSize: 100}})
	if err != nil {
		return nil, fmt.Errorf("list units: %w", err)
	}
	units, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[unit.Unit], error) {
		return svc.List(ctx, &unit.ListOptions{ListOptions: pagination.ListOptions{Page: pageNum, PageSize: 100}})
	})
	if err != nil {
		return nil, err
	}
	unitsByName := make(map[string]int, len(units))
	var baseUnitName string
	for _, u := range units {
		unitsByName[strings.ToLower(u.Name)] = u.ID
		if u.ID == baseUnit {
			baseUnitName = strings.ToLower(u.Name)
		}
	}

	// Heuristic targets by base unit kind.
	type convPlan struct {
		targetName string
		baseAmount float64
		convAmount float64
	}
	var plans []convPlan
	lower := baseUnitName
	switch {
	case strings.Contains(lower, "gram") || lower == "g":
		plans = []convPlan{
			{"ounce", 28.3495, 1},
			{"pound", 453.592, 1},
			{"kilogram", 1000, 1},
		}
	case strings.Contains(lower, "millilitre") || strings.Contains(lower, "milliliter") || strings.Contains(lower, "ml"):
		plans = []convPlan{
			{"cup", 236.588, 1},
			{"tablespoon", 14.787, 1},
			{"teaspoon", 4.929, 1},
		}
	default:
		return &AutoConversionsResult{
			FoodID:       foodID,
			BaseUnit:     baseUnit,
			Skipped:      []string{fmt.Sprintf("no conversion heuristics for base unit %q", baseUnitName)},
			ActionsTaken: []string{},
		}, nil
	}

	// Existing conversions for this food.
	csvc := c.UnitConversions()
	foodPtr := foodID
	convPage, err := csvc.List(ctx, &unit.ConversionListOptions{
		ListOptions: &pagination.ListOptions{PageSize: 100},
		FoodID:      &foodPtr,
	})
	if err != nil {
		return nil, fmt.Errorf("list conversions: %w", err)
	}
	exists := func(base, conv int) bool {
		for _, cv := range convPage.Results {
			if cv.BaseUnit != nil && cv.ConvertedUnit != nil && cv.Food != nil &&
				cv.BaseUnit.ID == base && cv.ConvertedUnit.ID == conv && cv.Food.ID == foodID {
				return true
			}
		}
		return false
	}

	res := &AutoConversionsResult{FoodID: foodID, BaseUnit: baseUnit, ActionsTaken: []string{}}
	for _, p := range plans {
		targetID, ok := unitsByName[p.targetName]
		if !ok {
			res.Skipped = append(res.Skipped, fmt.Sprintf("no unit named %q on instance", p.targetName))
			continue
		}
		if exists(baseUnit, targetID) {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s → %s already exists", baseUnitName, p.targetName))
			continue
		}
		if dryRun {
			res.Created = append(res.Created, unit.Conversion{
				BaseAmount:      p.baseAmount,
				BaseUnit:        &unit.Ref{ID: baseUnit},
				ConvertedAmount: p.convAmount,
				ConvertedUnit:   &unit.Ref{ID: targetID},
				Food:            &unit.FoodRef{ID: foodID},
			})
			res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("would_create_%s_to_%s", baseUnitName, p.targetName))
			continue
		}
		created, err := csvc.Create(ctx, &unit.Conversion{
			Name:            fmt.Sprintf("%s to %s", baseUnitName, p.targetName),
			BaseAmount:      p.baseAmount,
			BaseUnit:        &unit.Ref{ID: baseUnit},
			ConvertedAmount: p.convAmount,
			ConvertedUnit:   &unit.Ref{ID: targetID},
			Food:            &unit.FoodRef{ID: foodID},
		})
		if err != nil {
			return res, fmt.Errorf("create conversion %s → %s: %w", baseUnitName, p.targetName, err)
		}
		res.Created = append(res.Created, *created)
		res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("created_%s_to_%s", baseUnitName, p.targetName))
	}
	return res, nil
}
