package recipebook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestRecipeBookListOptions_Values(t *testing.T) {
	v := ListOptions{SpaceID: 1}.Values()
	assert.Equal(t, "1", v.Get("space"))
}

func TestRecipeBookListOptions_Values_Zero(t *testing.T) {
	assert.Empty(t, ListOptions{}.Values())
}

func TestRecipeBookListOptions_QueryString(t *testing.T) {
	q := ListOptions{
		SpaceID:     1,
		ListOptions: pagination.ListOptions{Search: "baking"},
	}.QueryString()
	assert.Contains(t, q, "space=1")
	assert.Contains(t, q, "query=baking")
}

func TestRecipeBookListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{SpaceID: 1}.ToPaginationOptions()
	assert.Equal(t, "1", p.Extra["space"])
}

func TestRecipeBookListOptions_ToPaginationOptions_MergesExtra(t *testing.T) {
	p := ListOptions{
		SpaceID:     1,
		ListOptions: pagination.ListOptions{Extra: map[string]string{"shared": "kept"}},
	}.ToPaginationOptions()
	assert.Equal(t, "1", p.Extra["space"])
	assert.Equal(t, "kept", p.Extra["shared"])
}
