package auditfood

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/swedishborgie/go-tandoor/unit"
)

// gateMacros is a foodNutrients payload that passes the 4-macro gate.
func gateMacros() []map[string]any {
	return []map[string]any{
		{"number": 1008, "amount": 158.0},
		{"number": 1003, "amount": 5.8},
		{"number": 1004, "amount": 1.0},
		{"number": 1005, "amount": 30.9},
	}
}

func iPtr(i int) *int { return &i }

func containsSubstr(haystack []string, sub string) bool {
	for _, s := range haystack {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// --- Prepare ---

func TestPrepare_CandidateMode_NoWrite(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)
	fdc := newFakeFDC(t, []map[string]any{
		{"fdcId": 2001, "dataType": "Foundation", "description": "Pasta, cooked", "score": 99.0, "foodNutrients": gateMacros()},
		{"fdcId": 1001, "dataType": "SR Legacy", "description": "PASTA, spaghetti, cooked", "score": 50.0, "foodNutrients": gateMacros()},
		{"fdcId": 3001, "dataType": "Branded", "description": "PastaBrand", "score": 999.0,
			"foodNutrients": []map[string]any{{"number": 1008, "amount": 0.0}}},
	})

	res, err := Prepare(context.Background(), c, fdc, &PrepareOptions{Name: "Pasta"})
	require.NoError(t, err)

	require.Equal(t, "candidates", res.Status)
	require.Empty(t, f.foods, "candidate mode must not write")
	require.Len(t, res.Candidates, 3)
	// SR Legacy outranks Foundation, which outranks Branded — score only
	// breaks ties within a data type.
	assert.Equal(t, 1001, res.Candidates[0].FDCID)
	assert.Equal(t, 2001, res.Candidates[1].FDCID)
	assert.Equal(t, 3001, res.Candidates[2].FDCID)
	assert.True(t, res.Candidates[0].MacroOK)
	assert.True(t, res.Candidates[1].MacroOK)
	assert.False(t, res.Candidates[2].MacroOK, "zero-energy branded food must fail the gate")
	assert.InDelta(t, 158.0, res.Candidates[0].Macros["calories"], 1e-9)
}

func TestPrepare_CreateWithFDCID(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(7)}
	f.propTypes = []property.Type{
		{ID: 10, Name: "Energy", Unit: "kcal", FDCID: iPtr(2085)},
		{ID: 11, Name: "Protein", Unit: "g", FDCID: iPtr(203)},
	}
	c := f.start(t)
	fdc := newFakeFDC(t, nil)

	res, err := Prepare(context.Background(), c, fdc, &PrepareOptions{
		Name: "Pasta, Cooked", FDCID: iPtr(172828), PluralName: "Pastas",
	})
	require.NoError(t, err)

	assert.Equal(t, "prepared", res.Status)
	assert.True(t, res.Created)
	assert.Equal(t, 172828, res.FDCID)
	require.Len(t, res.Attached, 2, "canned FDC food has two matching nutrients")
	require.Len(t, f.foods, 1)
	assert.Equal(t, "Pasta, Cooked", f.foods[0].Name)
	require.NotNil(t, f.foods[0].FDCID)
	assert.Equal(t, 172828, *f.foods[0].FDCID)
	require.Len(t, f.foods[0].Properties, 2)
}

func TestPrepare_ExistingSetsFDCIDAndAttaches(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(7)}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy", Unit: "kcal", FDCID: iPtr(2085)}}
	existing := f.addFood("Pasta")
	c := f.start(t)
	fdc := newFakeFDC(t, nil)

	res, err := Prepare(context.Background(), c, fdc, &PrepareOptions{
		Name: "Pasta", FDCID: iPtr(172828),
	})
	require.NoError(t, err)

	assert.Equal(t, "prepared", res.Status)
	assert.False(t, res.Created)
	require.Len(t, f.foods, 1, "must reuse, not create")
	require.NotNil(t, f.foods[0].FDCID)
	assert.Equal(t, 172828, *f.foods[0].FDCID)
	require.NotNil(t, existing)
	assert.Len(t, f.foods[0].Properties, 1)
}

func TestPrepare_ExistingWithoutFDCID_ReturnsCandidates(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Pasta")
	c := f.start(t)
	fdc := newFakeFDC(t, []map[string]any{
		{"fdcId": 1001, "dataType": "SR Legacy", "description": "PASTA, spaghetti, cooked", "foodNutrients": gateMacros()},
	})

	res, err := Prepare(context.Background(), c, fdc, &PrepareOptions{Name: "Pasta"})
	require.NoError(t, err)

	assert.Equal(t, "candidates", res.Status)
	require.NotNil(t, res.Food)
	require.Len(t, res.Candidates, 1)
	assert.Empty(t, f.foods[0].Properties, "candidate mode must not attach")
}

func TestPrepare_MissingWithoutFDCIDAndNoClient(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)

	_, err := Prepare(context.Background(), c, nil, &PrepareOptions{Name: "Pasta"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FDC API key")
}

// --- AttachMany ---

func TestAttachMany_CreatesAndUpdates(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(7)}
	f.propTypes = []property.Type{
		{ID: 10, Name: "Energy", Unit: "kcal"},
		{ID: 11, Name: "Protein", Unit: "g"},
		{ID: 12, Name: "Fiber", Unit: "g"},
	}
	existing := f.addFood("Pasta")
	amount := 150.0
	existing.Properties = []property.Property{{Type: property.Type{ID: 10, Name: "Energy", Unit: "kcal"}, PropertyAmount: &amount}}
	c := f.start(t)

	res, err := AttachMany(context.Background(), c, &AttachManyOptions{
		FoodID: existing.ID,
		Properties: []ManyProperty{
			{PropertyTypeName: "energy", Amount: 200.0}, // updates existing (case-insensitive)
			{PropertyTypeID: 11, Amount: 6.0},           // creates
			{PropertyTypeName: "Fiber", Amount: 2.5},    // creates by name
		},
	})
	require.NoError(t, err)

	require.Len(t, res.Attached, 3)
	assert.Equal(t, "updated", res.Attached[0].Action)
	assert.Equal(t, "created", res.Attached[1].Action)
	assert.Equal(t, "created", res.Attached[2].Action)
	require.Len(t, f.foods[0].Properties, 3)
	assert.InDelta(t, 200.0, *f.foods[0].Properties[0].PropertyAmount, 1e-9)
}

func TestAttachMany_DuplicateLastWins(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(7)}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy", Unit: "kcal"}}
	existing := f.addFood("Pasta")
	c := f.start(t)

	res, err := AttachMany(context.Background(), c, &AttachManyOptions{
		FoodID: existing.ID,
		Properties: []ManyProperty{
			{PropertyTypeName: "Energy", Amount: 1.0},
			{PropertyTypeID: 10, Amount: 2.0},
		},
	})
	require.NoError(t, err)
	require.Len(t, res.Attached, 1)
	assert.InDelta(t, 2.0, res.Attached[0].Amount, 1e-9)
}

func TestAttachMany_UnknownName(t *testing.T) {
	f := newFakeTandoor()
	f.propTypes = []property.Type{{ID: 10, Name: "Energy", Unit: "kcal"}}
	existing := f.addFood("Pasta")
	c := f.start(t)

	_, err := AttachMany(context.Background(), c, &AttachManyOptions{
		FoodID:     existing.ID,
		Properties: []ManyProperty{{PropertyTypeName: "Vitamin Z", Amount: 1.0}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no property type named")
}

func TestAttachMany_Empty(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)
	_, err := AttachMany(context.Background(), c, &AttachManyOptions{FoodID: 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one property")
}

// --- RecipeAudit ---

func TestRecipeAudit_FullReport(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{
		gramUnit(7),
		{ID: 8, Name: "ounce", BaseUnit: "gram"},
		{ID: 9, Name: "pinch"},
	}
	f.propTypes = []property.Type{
		{ID: 10, Name: "Energy", Unit: "kcal"},
		{ID: 11, Name: "Protein", Unit: "g"},
	}
	pasta := f.addFood("Pasta")
	amount := 150.0
	pasta.Properties = []property.Property{{Type: property.Type{ID: 10, Name: "Energy", Unit: "kcal"}, PropertyAmount: &amount}}
	basil := f.addFood("Basil")
	_ = basil
	f.convs = []unit.Conversion{{
		ID: 1, BaseUnit: &unit.Ref{ID: 7, Name: "gram"}, ConvertedUnit: &unit.Ref{ID: 8, Name: "ounce"},
	}}
	f.setRecipe(1, map[string]any{
		"id": 1, "name": "Pasta", "servings": 4,
		"steps": []map[string]any{{
			"id": 1,
			"ingredients": []map[string]any{
				{"id": 10, "food": map[string]any{"id": 1, "name": "Pasta"}, "unit": map[string]any{"id": 8, "name": "ounce"}, "amount": 2},
				{"id": 11, "food": map[string]any{"id": 2, "name": "Basil"}, "unit": map[string]any{"id": 9, "name": "pinch"}, "amount": 1},
			},
		}},
		"food_properties": []map[string]any{
			{"type": 10, "total_value": 632.4, "food_values": map[string]any{"1": 632.4}},
			{"type": 11, "total_value": nil, "food_values": map[string]any{}},
			{"type": 10, "total_value": 0, "food_values": map[string]any{}},
			{"type": 11, "total_value": 0, "food_values": map[string]any{"1": 0.0}},
		},
	})
	f.ingredients = []map[string]any{{
		"id": 10, "food": map[string]any{"id": 1},
		"used_in_recipes": []map[string]any{{"id": 1, "name": "Pasta"}, {"id": 5, "name": "Other"}},
	}}
	c := f.start(t)

	res, err := RecipeAudit(context.Background(), c, 1)
	require.NoError(t, err)
	assert.Equal(t, "Pasta", res.Recipe.Name)
	assert.Equal(t, 4, res.Servings)

	require.Len(t, res.Foods, 2)
	p := res.Foods[0]
	assert.Equal(t, "Pasta", p.Name)
	assert.Equal(t, []string{"Protein"}, p.MissingProperties)
	assert.Len(t, p.Conversions, 1)
	assert.Empty(t, p.MissingConversions)
	assert.Equal(t, 1, p.IngredientCount)
	assert.True(t, containsSubstr(p.Issues, "[no_fdc_id]"), "food without fdc_id must be flagged: %v", p.Issues)

	b := res.Foods[1]
	assert.Equal(t, "Basil", b.Name)
	assert.ElementsMatch(t, []string{"Energy", "Protein"}, b.MissingProperties)
	assert.Equal(t, []string{"pinch"}, b.MissingConversions)

	require.Len(t, res.FoodProperties, 4)
	assert.Equal(t, "Energy", res.FoodProperties[0].Name)
	require.NotNil(t, res.FoodProperties[0].TotalValue)
	assert.InDelta(t, 632.4, *res.FoodProperties[0].TotalValue, 1e-9)
	require.NotNil(t, res.FoodProperties[0].PerServing)
	assert.InDelta(t, 158.1, *res.FoodProperties[0].PerServing, 0.001)
	assert.False(t, res.FoodProperties[0].MissingValue)
	assert.True(t, res.FoodProperties[1].MissingValue)
	assert.True(t, res.FoodProperties[2].SuspiciousZero)
	assert.False(t, res.FoodProperties[2].MissingValue)
	assert.False(t, res.FoodProperties[3].SuspiciousZero)

	assert.Equal(t, []int{5}, p.OtherRecipes, "shared usage excludes the audited recipe")
}

func TestAsFloat(t *testing.T) {
	got, ok := asFloat(int(3))
	assert.True(t, ok)
	assert.InDelta(t, 3.0, got, 1e-9)
	got, ok = asFloat(json.Number("2.5"))
	assert.True(t, ok)
	assert.InDelta(t, 2.5, got, 1e-9)
	got, ok = asFloat("1.5")
	assert.True(t, ok)
	assert.InDelta(t, 1.5, got, 1e-9)
	_, ok = asFloat(nil)
	assert.False(t, ok)
	_, ok = asFloat("nope")
	assert.False(t, ok)
}

func TestFetchUnitsAndGramID(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{{ID: 1, Name: "g", BaseUnit: "gram"}, {ID: 2, Name: "cup", BaseUnit: "cups"}, {ID: 3, Name: "oz"}}
	c := f.start(t)

	units, err := FetchAllUnits(context.Background(), c)
	require.NoError(t, err)
	assert.Len(t, units, 3)

	id, err := FindGramUnitID(context.Background(), c)
	require.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestPrepare_ExistingWithFDCID_NoOverride(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(7)}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy", Unit: "kcal", FDCID: iPtr(2085)}}
	existing := f.addFood("Pasta")
	existing.FDCID = iPtr(999)
	c := f.start(t)
	fdc := newFakeFDC(t, nil)

	res, err := Prepare(context.Background(), c, fdc, &PrepareOptions{Name: "Pasta"})
	require.NoError(t, err)
	assert.Equal(t, "prepared", res.Status)
	assert.False(t, res.Created)
	assert.Equal(t, 999, res.FDCID, "must use the food's existing FDC ID")
}

func TestPrepare_NoAutoConversions(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(7)}
	c := f.start(t)
	fdc := newFakeFDC(t, nil)
	off := false

	res, err := Prepare(context.Background(), c, fdc, &PrepareOptions{Name: "Pasta", FDCID: iPtr(172828), AutoConversions: &off})
	require.NoError(t, err)
	assert.Equal(t, "prepared", res.Status)
	assert.Empty(t, f.convs, "conversions must be skipped when disabled")
	for _, a := range res.ActionsTaken {
		assert.NotContains(t, a, "conversion")
	}
}

func TestPrepare_EmptyName(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)
	_, err := Prepare(context.Background(), c, nil, &PrepareOptions{Name: " "})
	require.Error(t, err)
}

func TestPrepare_CandidateLimitClampAndRanks(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)
	// SR Legacy first; "Survey (FNDDS)" and bare "survey" tie on rank and
	// fall back to score; unknown data types sink to the bottom.
	fdc := newFakeFDC(t, []map[string]any{
		{"fdcId": 1, "dataType": "SR Legacy", "description": "Legacy", "score": 1.0, "foodNutrients": gateMacros()},
		{"fdcId": 2, "dataType": "Survey (FNDDS)", "description": "Survey Suffix", "score": 50.0, "foodNutrients": gateMacros()},
		{"fdcId": 3, "dataType": "survey", "description": "Survey Bare", "score": 40.0, "foodNutrients": gateMacros()},
		{"fdcId": 4, "dataType": "ND", "description": "Unknown Type", "score": 10.0, "foodNutrients": gateMacros()},
		// nil amount: the 4-macro gate must fail.
		{"fdcId": 5, "dataType": "Branded", "description": "Missing Protein", "score": 5.0, "foodNutrients": []map[string]any{{"number": 1008, "amount": 300.0}, {"number": 1004, "amount": 1.0}, {"number": 1005, "amount": 2.0}, {"number": 1003, "amount": nil}}},
	})

	res, err := Prepare(context.Background(), c, fdc, &PrepareOptions{Name: "Pasta", CandidateLimit: 99})
	require.NoError(t, err)
	require.Len(t, res.Candidates, 5)
	ids := []int{}
	for _, x := range res.Candidates {
		ids = append(ids, x.FDCID)
	}
	assert.Equal(t, []int{1, 2, 3, 5, 4}, ids)
	assert.False(t, res.Candidates[3].MacroOK, "nil-amount protein must fail the gate")
	assert.Empty(t, f.foods, "candidate mode must not write")
}

func TestPrepare_ListError(t *testing.T) {
	c, err := tandoor.NewClient("http://127.0.0.1:1") // unreachable
	require.NoError(t, err)
	fdc := newFakeFDC(t, nil)
	_, err = Prepare(context.Background(), c, fdc, &PrepareOptions{Name: "Pasta"})
	require.Error(t, err)
}

func TestAttachMany_Errors(t *testing.T) {
	f := newFakeTandoor()
	f.propTypes = []property.Type{{ID: 10, Name: "Energy", Unit: "kcal"}}
	c := f.start(t)

	_, err := AttachMany(context.Background(), c, &AttachManyOptions{FoodID: 1, Properties: []ManyProperty{{PropertyTypeID: 99, Amount: 1}}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown property_type_id")

	_, err = AttachMany(context.Background(), c, &AttachManyOptions{FoodID: 1, Properties: []ManyProperty{{Amount: 1}}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "set property_type_id or property_type_name")

	_, err = AttachMany(context.Background(), c, &AttachManyOptions{FoodID: 1, Properties: []ManyProperty{{PropertyTypeID: 10, Amount: 1}}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get food 1")
}

func TestRecipeAudit_NotFound(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)
	_, err := RecipeAudit(context.Background(), c, 99)
	require.Error(t, err)
}
