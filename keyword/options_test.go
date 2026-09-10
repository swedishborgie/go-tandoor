package keyword

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestKeywordListOptions_Values(t *testing.T) {
	opts := ListOptions{RecipeBookID: 5}
	v := opts.Values()
	assert.Equal(t, "5", v.Get("recipe_book"))
}

func TestKeywordListOptions_QueryString(t *testing.T) {
	q := ListOptions{
		RecipeBookID: 4,
		ListOptions:  pagination.ListOptions{Search: "veg"},
	}.QueryString()
	assert.Contains(t, q, "recipe_book=4")
	assert.Contains(t, q, "query=veg")
}

func TestKeywordListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{RecipeBookID: 4}.ToPaginationOptions()
	assert.Equal(t, "4", p.Extra["recipe_book"])
}

func TestKeywordListOptions_ToPaginationOptions_MergesExtra(t *testing.T) {
	p := ListOptions{
		RecipeBookID: 4,
		ListOptions:  pagination.ListOptions{Extra: map[string]string{"shared": "kept"}},
	}.ToPaginationOptions()
	assert.Equal(t, "4", p.Extra["recipe_book"])
	assert.Equal(t, "kept", p.Extra["shared"])
}
