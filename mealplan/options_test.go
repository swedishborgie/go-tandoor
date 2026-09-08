package mealplan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanListOptions_Values(t *testing.T) {
	opts := ListOptions{SpaceID: 2, DateFrom: "2024-01-01", DateTo: "2024-01-31"}
	v := opts.Values()
	assert.Equal(t, "2", v.Get("space"))
	assert.Equal(t, "2024-01-01", v.Get("date_from"))
	assert.Equal(t, "2024-01-31", v.Get("date_to"))
}
