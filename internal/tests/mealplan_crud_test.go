//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/mealplan"
)

func TestIntegrationMealPlanCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a meal type for testing
	mt := &mealplan.MealType{
		Name:  "Integration Test Meal Type",
		Order: 1,
		Time:  "12:00:00",
		Color: "#ff0000",
	}
	createdMT, err := client.MealTypes().Create(ctx, mt)
	if err != nil {
		t.Logf("meal type create error: %v", err)
	}
	require.NoError(t, err)
	require.NotZero(t, createdMT.ID)
	defer client.MealTypes().Delete(ctx, createdMT.ID)
	mealTypeID := createdMT.ID
	t.Logf("created meal type id: %d", mealTypeID)

	from := time.Now().Add(24 * time.Hour).UTC()
	// Create a meal plan entry
	mp := &mealplan.MealPlan{
		Title:      "Integration Test Meal",
		MealTypeID: mealTypeID,
		FromDate:   from,
		ToDate:     from.Add(1 * time.Hour),
		Servings:   2,
	}
	created, err := client.MealPlans().Create(ctx, mp)
	if err != nil {
		t.Logf("create error: %v", err)
	}
	require.NoError(t, err)
	require.NotNil(t, created)
	t.Logf("created meal plan: %+v", created)
	require.NotZero(t, created.ID)
	require.Equal(t, "Integration Test Meal", created.Title)
	// List to verify
	list, err := client.MealPlans().List(ctx, nil)
	require.NoError(t, err)
	t.Logf("meal plans count after create: %d", len(list.Results))
	// Try get
	_, err = client.MealPlans().Get(ctx, created.ID)
	if err != nil {
		t.Logf("get error: %v", err)
		// Try raw
		var raw map[string]any
		err2 := client.DoJSON(ctx, "GET", fmt.Sprintf("api/meal-plan/%d/", created.ID), nil, &raw)
		t.Logf("raw get err: %v raw: %+v", err2, raw)
	}
	require.NoError(t, err)

	// Get
	got, err := client.MealPlans().Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	// Update title
	got.Title = "Integration Test Meal Updated"
	updated, err := client.MealPlans().Update(ctx, got)
	require.NoError(t, err)
	require.Equal(t, "Integration Test Meal Updated", updated.Title)

	// Delete
	err = client.MealPlans().Delete(ctx, updated.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = client.MealPlans().Get(ctx, updated.ID)
	require.Error(t, err)
}
