package mealplan

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlan_UnmarshalJSON_MealTypeInt(t *testing.T) {
	var mp MealPlan
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"meal_type":2}`), &mp))
	assert.Equal(t, 2, mp.MealTypeID)
	assert.Nil(t, mp.MealType)
}

func TestMealPlan_UnmarshalJSON_MealTypeObject(t *testing.T) {
	var mp MealPlan
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"meal_type":{"id":3,"name":"Lunch"}}`), &mp))
	assert.Equal(t, 3, mp.MealTypeID)
	require.NotNil(t, mp.MealType)
	assert.Equal(t, "Lunch", mp.MealType.Name)
}

func TestMealPlan_UnmarshalJSON_NoMealType(t *testing.T) {
	var mp MealPlan
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"title":"Pasta","servings":4}`), &mp))
	assert.Equal(t, 1, mp.ID)
	assert.Equal(t, "Pasta", mp.Title)
	assert.InDelta(t, 4.0, mp.Servings, 1e-9)
	assert.Zero(t, mp.MealTypeID)
	assert.Nil(t, mp.MealType)
}

func TestMealPlan_UnmarshalJSON_MealTypeInvalid(t *testing.T) {
	// Neither an int nor a MealType object — both decode "successfully" as
	// zero values; the raw value is simply not represented.
	var mp MealPlan
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"meal_type":"lunch"}`), &mp))
	assert.Zero(t, mp.MealTypeID)
	assert.Nil(t, mp.MealType)
}

func TestMealPlan_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var mp MealPlan
	require.Error(t, json.Unmarshal([]byte(`{nope`), &mp))
}
