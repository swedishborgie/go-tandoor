package food

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
