package auditfood

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/detector"
	"github.com/swedishborgie/go-tandoor/idref"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/swedishborgie/go-tandoor/unit"
)

// RecipeAuditFood is the per-food row of a RecipeAuditResult: the cleanup
// table (state, issues, conversions) plus the shared-food safety signal
// (other_recipes).
type RecipeAuditFood struct {
	ID                 int               `json:"id"`
	Name               string            `json:"name"`
	FDCID              *int              `json:"fdc_id"`
	PropertyCount      int               `json:"property_count"`
	MissingProperties  []string          `json:"missing_properties,omitempty"` // names of instance property types the food lacks
	Issues             []string          `json:"issues,omitempty"`             // detector issues (name problems, no FDC ID, ...)
	Conversions        []unit.Conversion `json:"conversions"`
	MissingConversions []string          `json:"missing_conversions,omitempty"` // unit names this recipe uses for the food with no conversion
	IngredientCount    int               `json:"ingredient_count"`
	OtherRecipes       []int             `json:"other_recipes,omitempty"` // recipes other than the audited one using this food
}

// RecipeAuditProperty is one entry of the recipe's food_properties with the
// verification flags computed client-side.
type RecipeAuditProperty struct {
	TypeID         int      `json:"type_id"`
	Name           string   `json:"name"`
	TotalValue     *float64 `json:"total_value"`
	PerServing     *float64 `json:"per_serving,omitempty"`
	MissingValue   bool     `json:"missing_value"`   // total_value absent/null
	SuspiciousZero bool     `json:"suspicious_zero"` // total is 0 with no contributing food values (missing property or conversion)
	FoodValues     any      `json:"food_values,omitempty"`
}

// RecipeAuditResult is the full report for one recipe: the recipe itself,
// a per-food cleanup table, and the macro verification data.
type RecipeAuditResult struct {
	Recipe         *recipe.Recipe        `json:"recipe"`
	Servings       int                   `json:"servings"`
	Foods          []RecipeAuditFood     `json:"foods"`
	FoodProperties []RecipeAuditProperty `json:"food_properties"`
}

// RecipeAudit builds the whole cleanup/verification picture for a recipe in
// one pass: per-food state (FDC ID, properties vs the instance's expected
// set, detector issues, unit conversions incl. missing ones, usage in
// other recipes) plus the recipe's food_properties with per-serving totals
// and missing-value flags. Read-only.
func RecipeAudit(ctx context.Context, c *tandoor.Client, recipeID int) (*RecipeAuditResult, error) {
	r, err := c.Recipes().Get(ctx, recipeID)
	if err != nil {
		return nil, fmt.Errorf("get recipe %d: %w", recipeID, err)
	}

	types, err := FetchAllPropertyTypes(ctx, c)
	if err != nil {
		return nil, err
	}
	typeNames := make(map[int]string, len(types))
	expected := make([]property.Type, 0, len(types))
	for _, t := range types {
		typeNames[t.ID] = t.Name
		expected = append(expected, t)
	}

	// Distinct foods in first-seen order, plus the units and ingredient
	// counts each food uses in this recipe.
	foodOrder := make([]int, 0)
	foodSet := make(map[int]struct{})
	foodUnits := make(map[int]map[int]string) // foodID -> unitID -> unit name
	foodCounts := make(map[int]int)           // foodID -> ingredient count
	for _, step := range r.Steps {
		for _, ing := range step.Ingredients {
			if ing.Food == nil {
				continue
			}
			if _, ok := foodSet[ing.Food.ID]; !ok {
				foodSet[ing.Food.ID] = struct{}{}
				foodOrder = append(foodOrder, ing.Food.ID)
			}
			foodCounts[ing.Food.ID]++
			if ing.Unit != nil && ing.Unit.ID != 0 {
				if foodUnits[ing.Food.ID] == nil {
					foodUnits[ing.Food.ID] = make(map[int]string)
				}
				if _, seen := foodUnits[ing.Food.ID][ing.Unit.ID]; !seen {
					foodUnits[ing.Food.ID][ing.Unit.ID] = ing.Unit.Name
				}
			}
		}
	}

	reg := detector.NewRegistry()
	foods := make([]RecipeAuditFood, 0, len(foodOrder))
	for _, foodID := range foodOrder {
		f, err := c.Foods().Get(ctx, foodID)
		if err != nil {
			return nil, fmt.Errorf("get food %d: %w", foodID, err)
		}
		row := RecipeAuditFood{
			ID:            foodID,
			Name:          f.Name,
			FDCID:         f.FDCID,
			PropertyCount: len(f.Properties),
			Conversions:   []unit.Conversion{},
		}

		attached := make(map[int]struct{}, len(f.Properties))
		for _, p := range f.Properties {
			attached[p.Type.ID] = struct{}{}
		}
		for _, t := range expected {
			if _, ok := attached[t.ID]; !ok {
				row.MissingProperties = append(row.MissingProperties, t.Name)
			}
		}

		detFood := detector.Food{
			Name:            f.Name,
			FDCID:           f.FDCID,
			PropertyTypeIDs: PropertyTypeIDs(f.Properties),
		}
		for _, iss := range reg.CheckAll(detFood) {
			row.Issues = append(row.Issues, fmt.Sprintf("[%s] %s (%s)", iss.Category, iss.Message, iss.Severity))
		}

		page, err := c.UnitConversions().List(ctx, &unit.ConversionListOptions{
			ListOptions: &pagination.ListOptions{Page: 1, PageSize: 100},
			FoodID:      &foodID,
		})
		if err != nil {
			return nil, fmt.Errorf("list conversions for food %d: %w", foodID, err)
		}
		row.Conversions = append(row.Conversions, page.Results...)
		for unitID, unitName := range foodUnits[foodID] {
			covered := false
			for _, cv := range row.Conversions {
				if cv.BaseUnit != nil && cv.BaseUnit.ID == unitID {
					covered = true
					break
				}
				if cv.ConvertedUnit != nil && cv.ConvertedUnit.ID == unitID {
					covered = true
					break
				}
			}
			if !covered {
				name := unitName
				if name == "" {
					name = fmt.Sprintf("unit %d", unitID)
				}
				row.MissingConversions = append(row.MissingConversions, name)
			}
		}

		row.IngredientCount = foodCounts[foodID]
		row.OtherRecipes = otherRecipes(ctx, c, foodID, recipeID)
		foods = append(foods, row)
	}

	return &RecipeAuditResult{
		Recipe:         r,
		Servings:       r.Servings,
		Foods:          foods,
		FoodProperties: parseFoodProperties(r.FoodProperties, typeNames, r.Servings),
	}, nil
}

// otherRecipes collects the other recipes that use this food (from the
// ingredient list's used_in_recipes). A lookup failure is non-fatal: the
// shared-food signal is a hint, not a requirement.
func otherRecipes(ctx context.Context, c *tandoor.Client, foodID, recipeID int) []int {
	page, err := c.Ingredients().List(ctx, &ingredient.ListOptions{
		ListOptions: pagination.ListOptions{Page: 1, PageSize: 100, Extra: map[string]string{"food": fmt.Sprintf("%d", foodID)}},
	})
	if err != nil {
		return []int{}
	}
	seen := make(map[int]struct{})
	other := []int{}
	for _, ing := range page.Results {
		for _, ref := range ing.UsedInRecipes {
			id := ref.ID
			if id == 0 || id == recipeID {
				continue
			}
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				other = append(other, id)
			}
		}
	}
	return other
}

// parseFoodProperties normalizes the recipe's raw food_properties into
// verification rows. Entries carry a type ref (id or object), a possibly
// null total_value, and a food_values map.
func parseFoodProperties(v any, typeNames map[int]string, servings int) []RecipeAuditProperty {
	rows := []RecipeAuditProperty{}
	entries, _ := v.([]any)
	for _, e := range entries {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		row := RecipeAuditProperty{TypeID: idref.AsAnyID(m["type"])}
		row.Name = typeNames[row.TypeID]

		if total, ok := asFloat(m["total_value"]); ok {
			row.TotalValue = &total
			if servings > 0 {
				per := total / float64(servings)
				row.PerServing = &per
			}
		} else {
			row.MissingValue = true
		}

		if fv, has := m["food_values"]; has {
			row.FoodValues = fv
			if row.TotalValue != nil && *row.TotalValue == 0 && isEmptyObject(fv) {
				row.SuspiciousZero = true
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func asFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		var f float64
		if _, err := fmt.Sscanf(x, "%f", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

func isEmptyObject(v any) bool {
	m, ok := v.(map[string]any)
	return ok && len(m) == 0
}
