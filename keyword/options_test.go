package keyword

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywordListOptions_Values(t *testing.T) {
	opts := ListOptions{RecipeBookID: 5}
	v := opts.Values()
	assert.Equal(t, "5", v.Get("recipe_book"))
}
