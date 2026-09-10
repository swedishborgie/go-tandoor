package ingredient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIngredientUnmarshalJSON(t *testing.T) {
	t.Run("food and unit as ints", func(t *testing.T) {
		var i Ingredient
		err := json.Unmarshal([]byte(`{"id":1,"food":7,"unit":3,"amount":2.5,"note":"x"}`), &i)
		require.NoError(t, err)
		assert.Equal(t, 1, i.ID)
		assert.Equal(t, 7, i.FoodID)
		assert.Equal(t, 3, i.UnitID)
		require.NotNil(t, i.Amount)
		assert.InDelta(t, 2.5, *i.Amount, 1e-9)
		assert.Nil(t, i.Food)
	})

	t.Run("food and unit as objects", func(t *testing.T) {
		var i Ingredient
		err := json.Unmarshal([]byte(`{"food":{"id":7,"name":"Pasta"},"unit":{"id":3,"name":"g"}}`), &i)
		require.NoError(t, err)
		assert.Equal(t, 7, i.FoodID)
		assert.Equal(t, 3, i.UnitID)
		require.NotNil(t, i.Food)
		assert.Equal(t, "Pasta", i.Food.Name)
		require.NotNil(t, i.Unit)
		assert.Equal(t, "g", i.Unit.Name)
	})

	t.Run("food as unparseable value falls through", func(t *testing.T) {
		var i Ingredient
		err := json.Unmarshal([]byte(`{"food":"not-a-number"}`), &i)
		require.NoError(t, err)
		assert.Nil(t, i.FoodID)
		assert.Nil(t, i.Food)
	})

	t.Run("invalid json", func(t *testing.T) {
		var i Ingredient
		err := json.Unmarshal([]byte(`{`), &i)
		assert.Error(t, err)
	})
}

func TestIngredientMarshalJSON(t *testing.T) {
	t.Run("Food object takes precedence", func(t *testing.T) {
		b, err := json.Marshal(Ingredient{Food: &Food{ID: 7}, Unit: &Unit{ID: 3}})
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(b, &m))
		assert.EqualValues(t, 7, m["food"])
		assert.EqualValues(t, 3, m["unit"])
	})

	t.Run("FoodID int is used when Food is nil", func(t *testing.T) {
		b, err := json.Marshal(Ingredient{FoodID: 9, UnitID: 4})
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(b, &m))
		assert.EqualValues(t, 9, m["food"])
		assert.EqualValues(t, 4, m["unit"])
	})

	t.Run("ForceRefs emits null for unset refs", func(t *testing.T) {
		b, err := json.Marshal(Ingredient{ForceRefs: true})
		require.NoError(t, err)
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(b, &m))
		assert.JSONEq(t, "null", string(m["food"]))
		assert.JSONEq(t, "null", string(m["unit"]))
	})

	t.Run("non-int FoodID is omitted without ForceRefs", func(t *testing.T) {
		b, err := json.Marshal(Ingredient{FoodID: "weird"})
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(b, &m))
		_, ok := m["food"]
		assert.False(t, ok)
	})

	t.Run("zero int FoodID falls back to ForceRefs", func(t *testing.T) {
		b, err := json.Marshal(Ingredient{FoodID: 0, ForceRefs: true})
		require.NoError(t, err)
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(b, &m))
		assert.JSONEq(t, "null", string(m["food"]))
	})
}
