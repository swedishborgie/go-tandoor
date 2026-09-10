package food

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestFoodListOptions_Values(t *testing.T) {
	opts := ListOptions{CategoryID: 3, UnitID: 7}
	v := opts.Values()
	assert.Equal(t, "3", v.Get("category"))
	assert.Equal(t, "7", v.Get("unit"))
}

func TestFoodListOptions_ToPaginationOptions(t *testing.T) {
	opts := ListOptions{CategoryID: 2}
	p := opts.ToPaginationOptions()
	assert.Equal(t, "2", p.Extra["category"])
}

func TestFoodListOptions_ToPaginationOptions_MergesExistingExtra(t *testing.T) {
	opts := ListOptions{
		CategoryID:  2,
		UnitID:      9,
		ListOptions: pagination.ListOptions{Extra: map[string]string{"shared": "kept"}},
	}
	p := opts.ToPaginationOptions()
	assert.Equal(t, "2", p.Extra["category"])
	assert.Equal(t, "9", p.Extra["unit"])
	assert.Equal(t, "kept", p.Extra["shared"])
}

func TestFoodListOptions_QueryString(t *testing.T) {
	opts := ListOptions{
		CategoryID:  3,
		ListOptions: pagination.ListOptions{Search: "rice", PageSize: 25},
	}
	q := opts.ToPaginationOptions().QueryString()
	assert.Contains(t, q, "category=3")
	assert.Contains(t, q, "query=rice")
	assert.Contains(t, q, "page_size=25")
}
