package mealplan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestMealPlanListOptions_Values(t *testing.T) {
	opts := ListOptions{SpaceID: 2, DateFrom: "2024-01-01", DateTo: "2024-01-31"}
	v := opts.Values()
	assert.Equal(t, "2", v.Get("space"))
	assert.Equal(t, "2024-01-01", v.Get("date_from"))
	assert.Equal(t, "2024-01-31", v.Get("date_to"))
}

func TestMealPlanListOptions_QueryString(t *testing.T) {
	q := ListOptions{
		SpaceID:  1,
		UserID:   2,
		DateFrom: "2025-01-01",
	}.QueryString()
	assert.Contains(t, q, "space=1")
	assert.Contains(t, q, "user=2")
	assert.Contains(t, q, "date_from=2025-01-01")
}

func TestMealPlanListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{SpaceID: 1, UserID: 2, DateFrom: "a", DateTo: "b"}.ToPaginationOptions()
	assert.Equal(t, "1", p.Extra["space"])
	assert.Equal(t, "2", p.Extra["user"])
	assert.Equal(t, "a", p.Extra["date_from"])
	assert.Equal(t, "b", p.Extra["date_to"])
}

func TestMealPlanListOptions_ToPaginationOptions_MergesExtra(t *testing.T) {
	p := ListOptions{
		SpaceID:     1,
		ListOptions: pagination.ListOptions{Extra: map[string]string{"shared": "kept"}},
	}.ToPaginationOptions()
	assert.Equal(t, "1", p.Extra["space"])
	assert.Equal(t, "kept", p.Extra["shared"])
}
