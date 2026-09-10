package step

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestStepListOptions_Values(t *testing.T) {
	v := ListOptions{RecipeID: 5}.Values()
	assert.Equal(t, "5", v.Get("recipe"))
}

func TestStepListOptions_Values_Zero(t *testing.T) {
	assert.Empty(t, ListOptions{}.Values())
}

func TestStepListOptions_QueryString(t *testing.T) {
	q := ListOptions{
		RecipeID:    5,
		ListOptions: pagination.ListOptions{PageSize: 10},
	}.QueryString()
	assert.Contains(t, q, "recipe=5")
	assert.Contains(t, q, "page_size=10")
}

func TestStepListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{RecipeID: 5}.ToPaginationOptions()
	assert.Equal(t, "5", p.Extra["recipe"])
}

func TestStepListOptions_ToPaginationOptions_Zero(t *testing.T) {
	p := ListOptions{}.ToPaginationOptions()
	assert.Empty(t, p.Extra)
}
