package ingredient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestIngredientListOptions_Values(t *testing.T) {
	v := ListOptions{FoodID: 3, RecipeID: 8, SpaceID: 1}.Values()
	assert.Equal(t, "3", v.Get("food"))
	assert.Equal(t, "8", v.Get("recipe"))
	assert.Equal(t, "1", v.Get("space"))
}

func TestIngredientListOptions_Values_Zero(t *testing.T) {
	v := ListOptions{}.Values()
	assert.Empty(t, v)
}

func TestIngredientListOptions_QueryString(t *testing.T) {
	q := ListOptions{
		RecipeID:    8,
		ListOptions: pagination.ListOptions{Search: "flour"},
	}.QueryString()
	assert.Contains(t, q, "recipe=8")
	assert.Contains(t, q, "query=flour")
}

func TestIngredientListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{FoodID: 3, RecipeID: 8, SpaceID: 1}.ToPaginationOptions()
	assert.Equal(t, "3", p.Extra["food"])
	assert.Equal(t, "8", p.Extra["recipe"])
	assert.Equal(t, "1", p.Extra["space"])
}

func TestIngredientListOptions_ToPaginationOptions_MergesExtra(t *testing.T) {
	p := ListOptions{
		FoodID:      3,
		ListOptions: pagination.ListOptions{Extra: map[string]string{"shared": "kept"}},
	}.ToPaginationOptions()
	assert.Equal(t, "3", p.Extra["food"])
	assert.Equal(t, "kept", p.Extra["shared"])
}
