package ingredient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFoodIDValue_Nil(t *testing.T) {
	i := &Ingredient{}
	assert.Nil(t, i.FoodIDValue())
}

func TestFoodIDValue_Set(t *testing.T) {
	i := &Ingredient{Food: &Food{ID: 7}}
	v := i.FoodIDValue()
	require.NotNil(t, v)
	assert.Equal(t, 7, *v)
}

func TestUnitIDValue_Nil(t *testing.T) {
	i := &Ingredient{}
	assert.Nil(t, i.UnitIDValue())
}

func TestUnitIDValue_Set(t *testing.T) {
	i := &Ingredient{Unit: &Unit{ID: 3}}
	v := i.UnitIDValue()
	require.NotNil(t, v)
	assert.Equal(t, 3, *v)
}
