package auditfood

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/swedishborgie/go-tandoor/unit"
)

// fakeTandoor is an in-memory Tandoor API server for auditfood tests.
type fakeTandoor struct {
	mu        sync.Mutex
	foods     []food.Food
	nextID    int
	units     []unit.Unit
	propTypes []property.Type
	convs     []unit.Conversion
	recipes   map[int]map[string]any
	// ingredients, when set, is returned by GET /api/ingredient/.
	ingredients []map[string]any
	// pageCap, when > 0, splits the food list into pages of this size.
	pageCap int
}

// setRecipe stores a canned recipe for GET /api/recipe/{id}/.
func (f *fakeTandoor) setRecipe(id int, recipe map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.recipes == nil {
		f.recipes = make(map[int]map[string]any)
	}
	f.recipes[id] = recipe
}

func newFakeTandoor() *fakeTandoor {
	return &fakeTandoor{nextID: 1}
}

func (f *fakeTandoor) addFood(name string) *food.Food {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.foods = append(f.foods, food.Food{ID: f.nextID, Name: name})
	f.nextID++
	return &f.foods[len(f.foods)-1]
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func pageResp[T any](all []T, page, pageSize int) map[string]any {
	results := make([]T, 0, len(all))
	start := (page - 1) * pageSize
	if start < len(all) {
		end := start + pageSize
		if end > len(all) {
			end = len(all)
		}
		results = all[start:end]
	}
	next := any(nil)
	if start+pageSize < len(all) {
		next = "http://fake/api/"
	}
	return map[string]any{"count": len(all), "next": next, "previous": nil, "results": results}
}

func (f *fakeTandoor) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		q := r.URL.Query()

		switch {
		case p == "/api/food/" && r.Method == http.MethodPost:
			var in food.Food
			_ = json.NewDecoder(r.Body).Decode(&in)
			f.mu.Lock()
			in.ID = f.nextID
			f.nextID++
			f.foods = append(f.foods, in)
			out := in
			f.mu.Unlock()
			writeJSON(w, out)

		case p == "/api/food/":
			f.mu.Lock()
			foods := f.foods
			capLimit := f.pageCap
			f.mu.Unlock()
			page, pageSize := 1, 50
			if v, err := atoi(q.Get("page")); err == nil && v > 0 {
				page = v
			}
			if v, err := atoi(q.Get("page_size")); err == nil && v > 0 {
				pageSize = v
			}
			if capLimit > 0 {
				pageSize = capLimit
			}
			writeJSON(w, pageResp(foods, page, pageSize))

		case strings.HasPrefix(p, "/api/food/") && strings.Contains(p, "/merge/"):
			// PUT /api/food/{src}/merge/{dst}/
			parts := strings.Split(strings.Trim(p, "/"), "/")
			src, _ := atoi(parts[2])
			dst, _ := atoi(parts[4])
			f.mu.Lock()
			found := false
			for i := range f.foods {
				if f.foods[i].ID == dst {
					writeJSON(w, f.foods[i])
					found = true
					break
				}
			}
			f.mu.Unlock()
			if found {
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"detail": "not found"})
			_ = src

		case strings.HasPrefix(p, "/api/food/") && r.Method == http.MethodGet:
			id, _ := atoi(strings.TrimSuffix(strings.TrimPrefix(p, "/api/food/"), "/"))
			f.mu.Lock()
			found := false
			for i := range f.foods {
				if f.foods[i].ID == id {
					out := f.foods[i]
					writeJSON(w, out)
					found = true
					break
				}
			}
			f.mu.Unlock()
			if found {
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"detail": "not found"})

		case strings.HasPrefix(p, "/api/food/") && r.Method == http.MethodPatch:
			id, _ := atoi(strings.TrimSuffix(strings.TrimPrefix(p, "/api/food/"), "/"))
			var in food.Food
			_ = json.NewDecoder(r.Body).Decode(&in)
			f.mu.Lock()
			found := false
			for i := range f.foods {
				if f.foods[i].ID == id {
					if in.Name != "" {
						f.foods[i].Name = in.Name
					}
					if in.FDCID != nil {
						f.foods[i].FDCID = in.FDCID
					}
					if len(in.Properties) > 0 {
						f.foods[i].Properties = in.Properties
					}
					if in.PropertiesFoodAmount != nil {
						f.foods[i].PropertiesFoodAmount = in.PropertiesFoodAmount
					}
					if in.PropertiesFoodUnit != nil {
						f.foods[i].PropertiesFoodUnit = in.PropertiesFoodUnit
					}
					writeJSON(w, f.foods[i])
					found = true
					break
				}
			}
			f.mu.Unlock()
			if found {
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"detail": "not found"})

		case p == "/api/property-type/":
			f.mu.Lock()
			types := f.propTypes
			f.mu.Unlock()
			writeJSON(w, pageResp(types, 1, 200))

		case strings.HasPrefix(p, "/api/property-type/"):
			id, _ := atoi(strings.TrimSuffix(strings.TrimPrefix(p, "/api/property-type/"), "/"))
			f.mu.Lock()
			found := false
			for i := range f.propTypes {
				if f.propTypes[i].ID == id {
					out := f.propTypes[i]
					writeJSON(w, out)
					found = true
					break
				}
			}
			f.mu.Unlock()
			if found {
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"detail": "not found"})

		case strings.HasPrefix(p, "/api/property/") && r.Method == http.MethodPatch:
			var in property.Property
			_ = json.NewDecoder(r.Body).Decode(&in)
			writeJSON(w, in)

		case p == "/api/unit/":
			f.mu.Lock()
			units := f.units
			f.mu.Unlock()
			writeJSON(w, pageResp(units, 1, 200))

		case p == "/api/unit-conversion/":
			if r.Method == http.MethodPost {
				var in unit.Conversion
				_ = json.NewDecoder(r.Body).Decode(&in)
				in.ID = len(f.convs) + 1
				f.convs = append(f.convs, in)
				writeJSON(w, in)
				return
			}
			writeJSON(w, pageResp(f.convs, 1, 200))

		case strings.HasPrefix(p, "/api/recipe/") && r.Method == http.MethodGet:
			id, _ := atoi(strings.TrimSuffix(strings.TrimPrefix(p, "/api/recipe/"), "/"))
			f.mu.Lock()
			rec, ok := f.recipes[id]
			f.mu.Unlock()
			if ok {
				writeJSON(w, rec)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"detail": "not found"})

		case p == "/api/ingredient/" && r.Method == http.MethodGet:
			f.mu.Lock()
			ings := f.ingredients
			f.mu.Unlock()
			if ings == nil {
				writeJSON(w, map[string]any{"count": 0, "results": []map[string]any{}})
				return
			}
			writeJSON(w, map[string]any{"count": len(ings), "results": ings})

		case strings.HasPrefix(p, "/api/ingredient/"):
			id, _ := atoi(strings.TrimSuffix(strings.TrimPrefix(p, "/api/ingredient/"), "/"))
			if id == 55 {
				writeJSON(w, map[string]any{"id": 55, "food": map[string]any{"id": 7, "name": "Pasta"}})
				return
			}
			if id == 56 {
				writeJSON(w, map[string]any{"id": 56, "food": nil})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"detail": "not found"})

		default:
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]any{"detail": fmt.Sprintf("no route %s %s", r.Method, p)})
		}
	}
}

func atoi(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func (f *fakeTandoor) start(t *testing.T) *tandoor.Client {
	t.Helper()
	server := httptest.NewServer(f.handler())
	t.Cleanup(server.Close)
	c, err := tandoor.NewClient(server.URL)
	require.NoError(t, err)
	return c
}

func newFakeFDC(t *testing.T, foods []map[string]any) *fdc.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/fdc/v1/foods/search") || r.URL.Path == "/v1/foods/search" {
			resp := map[string]any{
				"totalHits": len(foods), "currentPage": 1, "totalPages": 1,
				"foods": foods,
			}
			writeJSON(w, resp)
			return
		}
		if strings.Contains(r.URL.Path, "/v1/food/") {
			writeJSON(w, fdcFoodJSON)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	c, err := fdc.NewClient(fdc.WithAPIKey("test"), fdc.WithBaseURL(server.URL+"/fdc"))
	require.NoError(t, err)
	return c
}

var fdcFoodJSON = map[string]any{
	"fdcId": 172828, "dataType": "ND", "description": "Pasta, cooked",
	"foodNutrients": []map[string]any{
		{"id": 2085, "amount": 158.1, "nutrient": map[string]any{"id": 2085, "name": "Energy"}},
		{"id": 203, "amount": 5.75, "nutrient": map[string]any{"id": 203, "name": "Protein"}},
	},
}

func gramUnit(id int) unit.Unit { return unit.Unit{ID: id, Name: "gram"} }

// --- FindDuplicates ---

func TestFindDuplicates_GroupsSimilar(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Chicken Breast")
	f.addFood("chicken breast ")
	f.addFood("Pasta")
	c := f.start(t)

	res, err := FindDuplicates(context.Background(), c, 0.7)
	require.NoError(t, err)
	assert.Equal(t, 3, res.Scanned)
	require.Len(t, res.Groups, 1)
	require.Len(t, res.Groups[0].Foods, 2)
	assert.Equal(t, 1, res.Groups[0].Foods[0].ID)
	assert.Equal(t, 2, res.Groups[0].Foods[1].ID)
}

func TestFindDuplicates_NoGroups(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Chicken")
	f.addFood("Pasta")
	c := f.start(t)

	res, err := FindDuplicates(context.Background(), c, 0)
	require.NoError(t, err)
	assert.InDelta(t, DefaultDuplicateThreshold, res.Threshold, 1e-9)
	assert.Empty(t, res.Groups)
}

func TestFindDuplicates_ListError(t *testing.T) {
	c := tandoorClientErroring(t)
	_, err := FindDuplicates(context.Background(), c, 0)
	require.Error(t, err)
}

// --- Ensure ---

func TestEnsure_Found(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Pasta")
	c := f.start(t)

	report, err := Ensure(context.Background(), c, &EnsureOptions{Names: []string{"pasta"}}, false)
	require.NoError(t, err)
	assert.Equal(t, 1, report.Summary.Found)
	res := report.Results["pasta"]
	assert.Equal(t, "found", res.Status)
	require.NotNil(t, res.Food)
	assert.Equal(t, 1, res.Food.ID)
}

func TestEnsure_NearMatch_Ambiguous(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Chicken Breast")
	c := f.start(t)

	report, err := Ensure(context.Background(), c, &EnsureOptions{
		Names: []string{"chicken breast fillet"}, Threshold: 0.3,
	}, false)
	require.NoError(t, err)
	res := report.Results["chicken breast fillet"]
	assert.Equal(t, "ambiguous", res.Status)
	require.Len(t, res.NearMatches, 1)
}

func TestEnsure_ForceCreate_DryRun(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)

	report, err := Ensure(context.Background(), c, &EnsureOptions{
		Names: []string{"Tofu"}, ForceCreate: true,
	}, true)
	require.NoError(t, err)
	assert.Equal(t, 1, report.Summary.WouldCreate)
	res := report.Results["Tofu"]
	assert.Equal(t, "would_create", res.Status)
	assert.Equal(t, []string{"food_create_dry_run"}, res.ActionsTaken)
}

func TestEnsure_ForceCreate_Real(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)

	report, err := Ensure(context.Background(), c, &EnsureOptions{
		Names: []string{"Tofu"}, ForceCreate: true,
	}, false)
	require.NoError(t, err)
	assert.Equal(t, 1, report.Summary.Created)
	res := report.Results["Tofu"]
	assert.Equal(t, "created", res.Status)
	require.NotNil(t, res.Food)
	assert.Equal(t, 1, res.Food.ID)
}

func TestEnsure_NotFound_NoFDC(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)

	report, err := Ensure(context.Background(), c, &EnsureOptions{Names: []string{"Unobtainium"}}, false)
	require.NoError(t, err)
	assert.Equal(t, 1, report.Summary.NotFound)
}

func TestEnsure_WithFDC_Candidates(t *testing.T) {
	f := newFakeTandoor()
	c := f.start(t)
	fdcC := newFakeFDC(t, []map[string]any{
		{"fdcId": 100, "description": "Unobtainium, lab grade", "dataType": "SD"},
	})

	report, err := Ensure(context.Background(), c, &EnsureOptions{
		Names: []string{"Unobtainium"}, FDC: fdcC,
	}, false)
	require.NoError(t, err)
	res := report.Results["Unobtainium"]
	require.Len(t, res.FdcCandidates, 1)
	assert.Equal(t, 100, res.FdcCandidates[0].FDCID)
}

func TestEnsure_ListError(t *testing.T) {
	c := tandoorClientErroring(t)
	_, err := Ensure(context.Background(), c, &EnsureOptions{Names: []string{"x"}}, false)
	require.Error(t, err)
}

// --- FindByName ---

func TestFindByName_Paginates(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Other")
	f.addFood("Pasta")
	f.pageCap = 1 // force 2 pages
	c := f.start(t)

	matches, err := FindByName(context.Background(), c, "pasta", 1)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, 2, matches[0].ID)
}

// --- Inspect ---

func TestInspect(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1)}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy", FDCID: intPtr(2085)}}
	fo := f.addFood("  Diced   CHICKEN  ")
	fo.FDCID = nil
	c := f.start(t)
	fdcC := newFakeFDC(t, []map[string]any{
		{"fdcId": 100, "description": "Chicken", "dataType": "SD", "score": 0.9},
	})

	res, err := Inspect(context.Background(), c, fdcC, fo.ID, 0)
	require.NoError(t, err)
	assert.Equal(t, "Diced Chicken", res.SuggestedName)
	assert.Empty(t, res.Ingredients)
	require.Len(t, res.FDCMatches, 1)
	assert.Equal(t, 100, res.FDCMatches[0].FDCID)
	assert.InDelta(t, 0.9, res.FDCMatches[0].Score, 1e-9)
}

func TestInspect_NoFDCID_SkipsFDC(t *testing.T) {
	f := newFakeTandoor()
	fo := f.addFood("Pasta")
	fo.FDCID = nil
	c := f.start(t)

	res, err := Inspect(context.Background(), c, nil, fo.ID, 0)
	require.NoError(t, err)
	assert.Empty(t, res.FDCMatches)
}

func TestInspect_GetError(t *testing.T) {
	c := tandoorClientErroring(t)
	_, err := Inspect(context.Background(), c, nil, 1, 0)
	require.Error(t, err)
}

// --- Fix ---

func TestFix_AlreadyCanonical(t *testing.T) {
	f := newFakeTandoor()
	fo := f.addFood("Pasta")
	c := f.start(t)

	res, err := Fix(context.Background(), c, &FixOptions{FoodID: fo.ID}, false)
	require.NoError(t, err)
	assert.Equal(t, "Pasta", res.NewName)
	assert.Equal(t, []string{"name_already_canonical"}, res.ActionsTaken)
}

func TestFix_Rename(t *testing.T) {
	f := newFakeTandoor()
	fo := f.addFood("  Diced   CHICKEN  ")
	c := f.start(t)

	res, err := Fix(context.Background(), c, &FixOptions{FoodID: fo.ID}, false)
	require.NoError(t, err)
	assert.Equal(t, "Diced Chicken", res.NewName)
	assert.Equal(t, []string{"renamed"}, res.ActionsTaken)
	require.NotNil(t, res.Food)
	assert.Equal(t, "Diced Chicken", res.Food.Name)
}

func TestFix_Rename_DryRun(t *testing.T) {
	f := newFakeTandoor()
	fo := f.addFood("  Diced   CHICKEN  ")
	c := f.start(t)

	res, err := Fix(context.Background(), c, &FixOptions{FoodID: fo.ID}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"would_rename"}, res.ActionsTaken)
	assert.Nil(t, res.Food)
}

func TestFix_Collision_Merge(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Diced Chicken") // canonical target (id 1)
	fo := f.addFood("  Diced   CHICKEN  ")
	c := f.start(t)

	res, err := Fix(context.Background(), c, &FixOptions{FoodID: fo.ID, MergeOnCollision: true}, false)
	require.NoError(t, err)
	assert.True(t, res.Merged)
	require.NotNil(t, res.Collision)
	assert.Equal(t, 1, res.Collision.ID)
	assert.Equal(t, []string{"merged_into_1"}, res.ActionsTaken)
}

func TestFix_Collision_DryRun_Merge(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Diced Chicken")
	fo := f.addFood("  Diced   CHICKEN  ")
	c := f.start(t)

	res, err := Fix(context.Background(), c, &FixOptions{FoodID: fo.ID, MergeOnCollision: true}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"would_merge_into_1"}, res.ActionsTaken)
}

func TestFix_Collision_NoMerge_Fails(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Diced Chicken")
	fo := f.addFood("  Diced   CHICKEN  ")
	c := f.start(t)

	_, err := Fix(context.Background(), c, &FixOptions{FoodID: fo.ID, MergeOnCollision: false}, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collision")
}

func TestFix_Collision_DryRun_NoMerge(t *testing.T) {
	f := newFakeTandoor()
	f.addFood("Diced Chicken")
	fo := f.addFood("  Diced   CHICKEN  ")
	c := f.start(t)

	res, err := Fix(context.Background(), c, &FixOptions{FoodID: fo.ID, MergeOnCollision: false}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"would_fail_collision"}, res.ActionsTaken)
}

func TestFix_GetError(t *testing.T) {
	c := tandoorClientErroring(t)
	_, err := Fix(context.Background(), c, &FixOptions{FoodID: 1}, false)
	require.Error(t, err)
}

// --- ResolveTargetFoodID / FindUnitIDByName ---

func TestResolveTargetFoodID(t *testing.T) {
	c := tandoorClientErroring(t)

	_, err := ResolveTargetFoodID(context.Background(), c, 1, 2)
	require.Error(t, err)
	_, err = ResolveTargetFoodID(context.Background(), c, 0, 0)
	require.Error(t, err)
	id, err := ResolveTargetFoodID(context.Background(), c, 7, 0)
	require.NoError(t, err)
	assert.Equal(t, 7, id)

	// Ingredient resolution against a real server.
	f := newFakeTandoor()
	c2 := f.start(t)
	id, err = ResolveTargetFoodID(context.Background(), c2, 0, 55)
	require.NoError(t, err)
	assert.Equal(t, 7, id)

	_, err = ResolveTargetFoodID(context.Background(), c2, 0, 56)
	require.Error(t, err)
	_, err = ResolveTargetFoodID(context.Background(), c2, 0, 99)
	require.Error(t, err)
}

func TestFindUnitIDByName(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1), {ID: 2, Name: "Ounce"}}
	c := f.start(t)

	id, err := FindUnitIDByName(context.Background(), c, "Gram")
	require.NoError(t, err)
	assert.Equal(t, 1, id)

	_, err = FindUnitIDByName(context.Background(), c, "bbl")
	require.Error(t, err)
}

func TestFindGramUnitID(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{
		{ID: 1, Name: "cup"},
		{ID: 2, Name: "g", BaseUnit: "gram"},
		{ID: 3, Name: "Grams"},
	}
	c := f.start(t)

	// base_unit "gram" wins over a matching name (Tandoor's default gram
	// unit is named "g").
	id, err := findGramUnitID(context.Background(), c)
	require.NoError(t, err)
	assert.Equal(t, 2, id)

	// Name fallback when no unit carries the gram base_unit slug.
	f.units = []unit.Unit{{ID: 5, Name: "g"}, {ID: 6, Name: "cup"}}
	id, err = findGramUnitID(context.Background(), c)
	require.NoError(t, err)
	assert.Equal(t, 5, id)

	f.units = []unit.Unit{{ID: 7, Name: "cup"}}
	_, err = findGramUnitID(context.Background(), c)
	require.Error(t, err)
}

// TestAttach_ResolvesGUnit: a food with no properties_food_unit must still
// attach cleanly when the instance's gram unit is named "g".
func TestAttach_ResolvesGUnit(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{{ID: 9, Name: "g", BaseUnit: "gram"}}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}}
	fo := f.addFood("Pasta")
	c := f.start(t)

	res, err := Attach(context.Background(), c, &AttachOptions{
		FoodID: fo.ID, PropertyTypeID: 10, Amount: 158,
	}, false)
	require.NoError(t, err)
	assert.Equal(t, "created", res.Action)
	require.NotNil(t, res.Per100)
	assert.Equal(t, 9, res.Per100.UnitID)
}

// --- Attach ---

func TestAttach_Create(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1)}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}}
	fo := f.addFood("Pasta")
	c := f.start(t)

	res, err := Attach(context.Background(), c, &AttachOptions{
		FoodID: fo.ID, PropertyTypeID: 10, Amount: 158,
	}, false)
	require.NoError(t, err)
	assert.Equal(t, "created", res.Action)
	require.NotNil(t, res.Food)
	assert.Len(t, res.Food.Properties, 1)
	require.NotNil(t, res.Per100)
	assert.InDelta(t, 100.0, res.Per100.Amount, 1e-9)
	assert.Equal(t, 1, res.Per100.UnitID)
	assert.Equal(t, []string{"food_properties_patched"}, res.ActionsTaken)
}

func TestAttach_Update(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1)}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}}
	amount := 158.0
	fo := f.addFood("Pasta")
	fo.Properties = []property.Property{{ID: 99, Type: property.Type{ID: 10, Name: "Energy"}, PropertyAmount: &amount}}
	fo.PropertiesFoodAmount = &amount
	fo.PropertiesFoodUnit = 1
	c := f.start(t)

	res, err := Attach(context.Background(), c, &AttachOptions{
		FoodID: fo.ID, PropertyTypeID: 10, Amount: 200,
	}, false)
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Action)
	require.NotNil(t, res.Property)
	assert.InDelta(t, 200.0, *res.Property.PropertyAmount, 1e-9)
}

func TestAttach_DryRun(t *testing.T) {
	f := newFakeTandoor()
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}}
	fo := f.addFood("Pasta")
	c := f.start(t)

	res, err := Attach(context.Background(), c, &AttachOptions{
		FoodID: fo.ID, PropertyTypeID: 10,
	}, true)
	require.NoError(t, err)
	assert.Equal(t, "would_create", res.Action)
}

func TestAttach_DryRun_Existing(t *testing.T) {
	f := newFakeTandoor()
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}}
	amount := 1.0
	fo := f.addFood("Pasta")
	fo.Properties = []property.Property{{ID: 99, Type: property.Type{ID: 10}, PropertyAmount: &amount}}
	c := f.start(t)

	res, err := Attach(context.Background(), c, &AttachOptions{
		FoodID: fo.ID, PropertyTypeID: 10,
	}, true)
	require.NoError(t, err)
	assert.Equal(t, "would_update", res.Action)
}

func TestAttach_Errors(t *testing.T) {
	f := newFakeTandoor()
	fo := f.addFood("Pasta")
	c := f.start(t)

	_, err := Attach(context.Background(), c, &AttachOptions{FoodID: fo.ID}, false)
	require.Error(t, err)

	// Property type missing on server.
	_, err = Attach(context.Background(), c, &AttachOptions{FoodID: fo.ID, PropertyTypeID: 42}, false)
	require.Error(t, err)

	// Food missing on server.
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}}
	_, err = Attach(context.Background(), c, &AttachOptions{FoodID: 999, PropertyTypeID: 10}, false)
	require.Error(t, err)

	// No unit resolvable for per-100 basis.
	_, err = Attach(context.Background(), c, &AttachOptions{FoodID: fo.ID, PropertyTypeID: 10}, false)
	require.Error(t, err)
}

// --- AttachFDCProperties ---

func TestAttachFDCProperties(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1)}
	fdcID2085 := 2085
	f.propTypes = []property.Type{
		{ID: 10, Name: "Energy", FDCID: &fdcID2085},
		{ID: 11, Name: "Protein", FDCID: intPtr(203)},
	}
	fo := f.addFood("Pasta")
	fo.FDCID = intPtr(172828)
	c := f.start(t)
	fdcC := newFakeFDC(t, nil)

	res, err := AttachFDCProperties(context.Background(), c, fdcC, &FdcAttachOptions{FoodID: fo.ID}, false)
	require.NoError(t, err)
	assert.Equal(t, 172828, res.FDCID)
	require.Len(t, res.Attached, 2)
	assert.Equal(t, "created", res.Attached[0].Action)
	require.NotNil(t, res.Per100)
	assert.Contains(t, res.ActionsTaken, "food_properties_patched(2)")
}

func TestAttachFDCProperties_DryRun(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1)}
	fdcID2085 := 2085
	f.propTypes = []property.Type{{ID: 10, Name: "Energy", FDCID: &fdcID2085}}
	fo := f.addFood("Pasta")
	fo.FDCID = intPtr(172828)
	c := f.start(t)
	fdcC := newFakeFDC(t, nil)

	res, err := AttachFDCProperties(context.Background(), c, fdcC, &FdcAttachOptions{FoodID: fo.ID}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"would_patch_food_properties"}, res.ActionsTaken)
}

func TestAttachFDCProperties_Errors(t *testing.T) {
	f := newFakeTandoor()
	fo := f.addFood("Pasta")
	c := f.start(t)

	_, err := AttachFDCProperties(context.Background(), c, nil, &FdcAttachOptions{FoodID: fo.ID}, false)
	require.Error(t, err)

	fdcC := newFakeFDC(t, nil)
	// Food has no FDC id and none provided.
	_, err = AttachFDCProperties(context.Background(), c, fdcC, &FdcAttachOptions{FoodID: fo.ID}, false)
	require.Error(t, err)
}

func TestAttachFDCProperties_NoMatches(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1)}
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}} // no FDCID
	fo := f.addFood("Pasta")
	fo.FDCID = intPtr(172828)
	c := f.start(t)
	fdcC := newFakeFDC(t, nil)

	res, err := AttachFDCProperties(context.Background(), c, fdcC, &FdcAttachOptions{FoodID: fo.ID}, false)
	require.NoError(t, err)
	assert.Empty(t, res.Attached)
	require.Len(t, res.Skipped, 1)
}

// --- AutoConversions ---

func TestAutoConversions_Gram(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{
		gramUnit(1), {ID: 2, Name: "Ounce"}, {ID: 3, Name: "Pound"}, {ID: 4, Name: "Kilogram"},
	}
	fo := f.addFood("Pasta")
	fo.PropertiesFoodUnit = 1
	c := f.start(t)

	res, err := AutoConversions(context.Background(), c, fo.ID, false)
	require.NoError(t, err)
	assert.Equal(t, 1, res.BaseUnit)
	assert.Len(t, res.Created, 3)
}

func TestAutoConversions_Millilitre(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{
		{ID: 1, Name: "millilitre"}, {ID: 2, Name: "Cup"}, {ID: 3, Name: "Tablespoon"}, {ID: 4, Name: "Teaspoon"},
	}
	fo := f.addFood("Olive Oil")
	fo.PropertiesFoodUnit = 1
	c := f.start(t)

	res, err := AutoConversions(context.Background(), c, fo.ID, false)
	require.NoError(t, err)
	assert.Len(t, res.Created, 3)
}

func TestAutoConversions_UnknownBaseUnit(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{{ID: 1, Name: "pinch"}}
	fo := f.addFood("Salt")
	fo.PropertiesFoodUnit = 1
	c := f.start(t)

	res, err := AutoConversions(context.Background(), c, fo.ID, false)
	require.NoError(t, err)
	assert.Empty(t, res.Created)
	require.Len(t, res.Skipped, 1)
	assert.Contains(t, res.Skipped[0], "no conversion heuristics")
}

func TestAutoConversions_MissingTargets(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1)} // no ounce/pound/kilogram
	fo := f.addFood("Pasta")
	fo.PropertiesFoodUnit = 1
	c := f.start(t)

	res, err := AutoConversions(context.Background(), c, fo.ID, false)
	require.NoError(t, err)
	assert.Empty(t, res.Created)
	assert.Len(t, res.Skipped, 3)
}

func TestAutoConversions_Existing(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1), {ID: 2, Name: "Ounce"}, {ID: 3, Name: "Pound"}, {ID: 4, Name: "Kilogram"}}
	fo := f.addFood("Pasta")
	fo.PropertiesFoodUnit = 1
	f.convs = []unit.Conversion{{
		ID: 1, BaseUnit: &unit.Ref{ID: 1}, ConvertedUnit: &unit.Ref{ID: 2}, Food: &unit.FoodRef{ID: fo.ID},
	}}
	c := f.start(t)

	res, err := AutoConversions(context.Background(), c, fo.ID, false)
	require.NoError(t, err)
	assert.Len(t, res.Created, 2)
	require.Len(t, res.Skipped, 1)
	assert.Contains(t, res.Skipped[0], "already exists")
}

func TestAutoConversions_DryRun(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1), {ID: 2, Name: "Ounce"}, {ID: 3, Name: "Pound"}, {ID: 4, Name: "Kilogram"}}
	fo := f.addFood("Pasta")
	fo.PropertiesFoodUnit = 1
	c := f.start(t)

	res, err := AutoConversions(context.Background(), c, fo.ID, true)
	require.NoError(t, err)
	assert.Len(t, res.Created, 3)
	assert.Equal(t, "would_create_gram_to_ounce", res.ActionsTaken[0])
}

func TestAutoConversions_BaseUnitResolution(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{gramUnit(1), {ID: 2, Name: "Ounce"}}
	fo := f.addFood("Pasta") // no properties_food_unit → resolves "gram"
	c := f.start(t)

	res, err := AutoConversions(context.Background(), c, fo.ID, false)
	require.NoError(t, err)
	assert.Equal(t, 1, res.BaseUnit)
}

func TestAutoConversions_NoGramUnit(t *testing.T) {
	f := newFakeTandoor()
	f.units = []unit.Unit{{ID: 1, Name: "pinch"}}
	fo := f.addFood("Salt") // no base unit, no gram on instance
	c := f.start(t)

	_, err := AutoConversions(context.Background(), c, fo.ID, false)
	require.Error(t, err)
}

// --- Registry ---

func TestNewRegistry_Success(t *testing.T) {
	f := newFakeTandoor()
	f.propTypes = []property.Type{{ID: 10, Name: "Energy"}, {ID: 11, Name: "Protein"}}
	c := f.start(t)

	var warned []string
	reg := NewRegistry(context.Background(), c, func(format string, args ...any) {
		warned = append(warned, fmt.Sprintf(format, args...))
	})
	require.NotNil(t, reg)
	assert.Empty(t, warned)
}

func TestNewRegistry_WarnsOnFailure(t *testing.T) {
	c := tandoorClientErroring(t)
	var warned []string
	reg := NewRegistry(context.Background(), c, func(format string, args ...any) {
		warned = append(warned, fmt.Sprintf(format, args...))
	})
	require.NotNil(t, reg)
	require.Len(t, warned, 1)
	assert.Contains(t, warned[0], "could not enumerate property types")
}

func TestPropertyTypeIDs(t *testing.T) {
	amount := 1.0
	props := []property.Property{
		{Type: property.Type{ID: 3}},
		{Type: property.Type{ID: 7}, PropertyAmount: &amount},
	}
	assert.Equal(t, []int{3, 7}, PropertyTypeIDs(props))
	assert.Equal(t, []int{}, PropertyTypeIDs(nil))
}

// --- helpers ---

func intPtr(i int) *int { return &i }

// tandoorClientErroring returns a client pointed at a server that always 500s.
func tandoorClientErroring(t *testing.T) *tandoor.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]any{"detail": "boom"})
	}))
	t.Cleanup(server.Close)
	c, err := tandoor.NewClient(server.URL)
	require.NoError(t, err)
	return c
}
