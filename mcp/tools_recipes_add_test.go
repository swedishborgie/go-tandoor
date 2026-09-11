package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	tandoor "github.com/swedishborgie/go-tandoor"
)

// recipeFake is a small stateful fake: GET returns the stored recipe, PUT
// replaces it and records the last body so tests can assert on the exact
// payload the tool sent.
type recipeFake struct {
	recipe   map[string]any
	lastBody map[string]any
}

func (f *recipeFake) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/recipe/1/" {
			_ = json.NewEncoder(w).Encode(f.recipe)
			return
		}
		if r.Method == http.MethodPut && r.URL.Path == "/api/recipe/1/" {
			_ = json.NewDecoder(r.Body).Decode(&f.lastBody)
			f.recipe = f.lastBody // stateful: later GETs see the update
			_ = json.NewEncoder(w).Encode(f.recipe)
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/api/recipe/" {
			_ = json.NewDecoder(r.Body).Decode(&f.lastBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1})
			return
		}
		http.Error(w, `{"detail": "no route"}`, http.StatusNotFound)
	}
}

func newRecipeFakeClient(t *testing.T, fake *recipeFake) *client.Client {
	t.Helper()
	srv := httptest.NewServer(fake.handler())
	t.Cleanup(srv.Close)
	c, err := tandoor.NewClient(srv.URL)
	require.NoError(t, err)
	s := NewServer(c, nil)
	mc, err := client.NewInProcessClient(s)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, mc.Start(ctx))
	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "tandoor-mcp-test", Version: "0.0.0"}
	_, err = mc.Initialize(ctx, initReq)
	require.NoError(t, err)
	return mc
}

func stepsOf(t *testing.T, body map[string]any) []map[string]any {
	t.Helper()
	raw, ok := body["steps"].([]any)
	require.True(t, ok, "steps must be an array, got %T", body["steps"])
	out := make([]map[string]any, 0, len(raw))
	for _, s := range raw {
		m, ok := s.(map[string]any)
		require.True(t, ok)
		out = append(out, m)
	}
	return out
}

func TestRecipeAddIngredients_AppendsAndSkipsDuplicates(t *testing.T) {
	existing := map[string]any{"food_id": float64(2), "amount": float64(1), "unit_id": float64(7)}
	fake := &recipeFake{recipe: map[string]any{
		"id":   1,
		"name": "Pancakes",
		"steps": []any{
			map[string]any{"name": "Mix", "ingredients": []any{existing}},
			map[string]any{"name": "Cook", "ingredients": []any{}},
		},
	}}
	c := newRecipeFakeClient(t, fake)

	res := callTool(t, c, "recipe_add_ingredients", map[string]any{
		"recipe_id":   1,
		"ingredients": []any{map[string]any{"food_id": 5, "amount": 2, "unit_id": 7}},
	})
	require.False(t, res.IsError, resultText(t, res))
	text := resultText(t, res)
	require.Contains(t, text, `"added":1`)
	require.Contains(t, text, `"skipped":0`)

	steps := stepsOf(t, fake.lastBody)
	require.Len(t, steps, 2)
	ings, _ := steps[0]["ingredients"].([]any)
	require.Len(t, ings, 2, "new ingredient appended to step 0")
	require.Empty(t, steps[1]["ingredients"], "step 1 untouched")

	// Idempotent re-run: identical entry is skipped, nothing grows.
	res = callTool(t, c, "recipe_add_ingredients", map[string]any{
		"recipe_id":   1,
		"ingredients": []any{map[string]any{"food_id": 5, "amount": 2, "unit_id": 7}},
	})
	require.False(t, res.IsError, resultText(t, res))
	require.Contains(t, resultText(t, res), `"added":0`)
	require.Contains(t, resultText(t, res), `"skipped":1`)
	ings, _ = stepsOf(t, fake.lastBody)[0]["ingredients"].([]any)
	require.Len(t, ings, 2)
}

func TestRecipeAddIngredients_CreatesFirstStep(t *testing.T) {
	fake := &recipeFake{recipe: map[string]any{"id": 1, "name": "Pancakes"}}
	c := newRecipeFakeClient(t, fake)

	res := callTool(t, c, "recipe_add_ingredients", map[string]any{
		"recipe_id":   1,
		"ingredients": []any{map[string]any{"food_id": 5}},
	})
	require.False(t, res.IsError, resultText(t, res))

	steps := stepsOf(t, fake.lastBody)
	require.Len(t, steps, 1)
	require.Equal(t, "Instructions", steps[0]["name"])
	require.Len(t, steps[0]["ingredients"].([]any), 1)
}

func TestRecipeAddIngredients_Validation(t *testing.T) {
	fake := &recipeFake{recipe: map[string]any{"id": 1, "name": "Pancakes"}}
	c := newRecipeFakeClient(t, fake)

	// Empty ingredients.
	res := callTool(t, c, "recipe_add_ingredients", map[string]any{
		"recipe_id":   1,
		"ingredients": []any{},
	})
	require.True(t, res.IsError)

	// Entry without food_id.
	res = callTool(t, c, "recipe_add_ingredients", map[string]any{
		"recipe_id":   1,
		"ingredients": []any{map[string]any{"amount": 1}},
	})
	require.True(t, res.IsError)
	require.Contains(t, resultText(t, res), "food_id")

	// Non-object entry.
	res = callTool(t, c, "recipe_add_ingredients", map[string]any{
		"recipe_id":   1,
		"ingredients": []any{1},
	})
	require.True(t, res.IsError)

	// step_index out of range (recipe has no steps, only 0 is valid).
	res = callTool(t, c, "recipe_add_ingredients", map[string]any{
		"recipe_id":   1,
		"step_index":  1,
		"ingredients": []any{map[string]any{"food_id": 5}},
	})
	require.True(t, res.IsError)
	require.Contains(t, resultText(t, res), "step_index")
}

func TestRecipeAddStep(t *testing.T) {
	fake := &recipeFake{recipe: map[string]any{
		"id":    1,
		"name":  "Pancakes",
		"steps": []any{map[string]any{"name": "Mix", "ingredients": []any{}}},
	}}
	c := newRecipeFakeClient(t, fake)

	res := callTool(t, c, "recipe_add_step", map[string]any{
		"recipe_id":   1,
		"name":        "Cook",
		"ingredients": []any{map[string]any{"food_id": 9, "note": "butter"}},
	})
	require.False(t, res.IsError, resultText(t, res))

	steps := stepsOf(t, fake.lastBody)
	require.Len(t, steps, 2)
	require.Equal(t, "Cook", steps[1]["name"])
	ings, _ := steps[1]["ingredients"].([]any)
	require.Len(t, ings, 1)
	// No ingredients given → still sent as [] (Tandoor 400s on missing key).
	res = callTool(t, c, "recipe_add_step", map[string]any{
		"recipe_id": 1,
		"name":      "Serve",
	})
	require.False(t, res.IsError, resultText(t, res))
	steps = stepsOf(t, fake.lastBody)
	require.Len(t, steps, 3)
	_, has := steps[2]["ingredients"]
	require.True(t, has, "step without ingredients must carry an empty array")
}

func TestRecipeCreate_MergesTopLevelIngredients(t *testing.T) {
	fake := &recipeFake{recipe: map[string]any{"id": 1}}
	c := newRecipeFakeClient(t, fake)

	res := callTool(t, c, "recipe_create", map[string]any{
		"data": map[string]any{
			"name":  "Pancakes",
			"steps": []any{map[string]any{"name": "Mix"}}, // no ingredients key
			// Top-level list: Tandoor would drop this; we must not.
			"ingredients": []any{
				map[string]any{"food_id": 5, "amount": 2, "unit_id": 7},
				map[string]any{"food_id": 6, "no_amount": true},
			},
		},
	})
	require.False(t, res.IsError, resultText(t, res))

	body := fake.lastBody
	_, hasTop := body["ingredients"]
	require.False(t, hasTop, "top-level ingredients must be removed from the payload")
	steps := stepsOf(t, body)
	require.Len(t, steps, 1)
	ings, _ := steps[0]["ingredients"].([]any)
	require.Len(t, ings, 2, "top-level ingredients merged into the first step")
}

func TestRecipeCreate_CreatesStepForBareIngredients(t *testing.T) {
	fake := &recipeFake{recipe: map[string]any{"id": 1}}
	c := newRecipeFakeClient(t, fake)

	res := callTool(t, c, "recipe_create", map[string]any{
		"data": map[string]any{
			"name":        "Pancakes",
			"ingredients": []any{map[string]any{"food_id": 5}},
		},
	})
	require.False(t, res.IsError, resultText(t, res))
	steps := stepsOf(t, fake.lastBody)
	require.Len(t, steps, 1)
	require.Equal(t, "Instructions", steps[0]["name"])
	require.Len(t, steps[0]["ingredients"].([]any), 1)
}

func TestRecipeWrites_RejectSilentlyDroppedKeys(t *testing.T) {
	fake := &recipeFake{recipe: map[string]any{"id": 1}}
	c := newRecipeFakeClient(t, fake)

	// "source" is ignored by Tandoor; the field is "source_url".
	for _, tool := range []string{"recipe_create", "recipe_update", "recipe_patch"} {
		args := map[string]any{"data": map[string]any{"name": "x", "source": "blog"}}
		if tool != "recipe_create" {
			args["id"] = 1
		}
		res := callTool(t, c, tool, args)
		require.Truef(t, res.IsError, "%s with source must fail", tool)
		require.Contains(t, resultText(t, res), "source_url")
	}

	// Top-level ingredients on update/patch are a silent drop → reject.
	for _, tool := range []string{"recipe_update", "recipe_patch"} {
		res := callTool(t, c, tool, map[string]any{
			"id":   1,
			"data": map[string]any{"ingredients": []any{map[string]any{"food_id": 5}}},
		})
		require.Truef(t, res.IsError, "%s with top-level ingredients must fail", tool)
		require.Contains(t, resultText(t, res), "recipe_add_ingredients")
	}
}

func TestListAllEmptyReturnsEmptyArray(t *testing.T) {
	// The generic fake answers list paths with an empty envelope; all=true
	// must yield [] (not null) so jq and callers can count safely.
	c := newTestClient(t)
	res := callTool(t, c, "unit_conversion_list", map[string]any{"all": true})
	require.False(t, res.IsError, resultText(t, res))
	text := strings.TrimSpace(resultText(t, res))
	require.Equal(t, "[]", text)

	// Non-all shape carries an empty results array, not null.
	res = callTool(t, c, "unit_conversion_list", nil)
	require.False(t, res.IsError, resultText(t, res))
	require.Contains(t, resultText(t, res), `"results":[]`)
}
