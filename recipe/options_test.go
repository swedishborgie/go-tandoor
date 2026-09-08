package recipe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecipeListOptions_Values(t *testing.T) {
	trueVal := true
	opts := ListOptions{
		KeywordIDs:   []int{1, 2, 3},
		SpaceID:      5,
		RecipeBookID: 7,
		IsFavorite:   &trueVal,
	}
	v := opts.Values()
	assert.Equal(t, "1,2,3", v.Get("keyword"))
	assert.Equal(t, "5", v.Get("space"))
	assert.Equal(t, "7", v.Get("recipe_book"))
	assert.Equal(t, "true", v.Get("is_favorite"))
}

func TestRecipeListOptions_ToPaginationOptions(t *testing.T) {
	opts := ListOptions{
		KeywordIDs: []int{4, 5},
		SpaceID:    9,
	}
	p := opts.ToPaginationOptions()
	assert.NotNil(t, p)
	assert.Equal(t, "4,5", p.Extra["keyword"])
	assert.Equal(t, "9", p.Extra["space"])
}
