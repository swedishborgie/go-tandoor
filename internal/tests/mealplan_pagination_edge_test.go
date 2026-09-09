//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/mealplan"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestIntegrationMealPlanPaginationEdge(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a meal type for the plans
	mt := &mealplan.MealType{
		Name:  "Edge MealType",
		Order: 1,
		Time:  strPtr("18:00:00"),
		Color: strPtr("#abcdef"),
	}
	createdMT, err := client.MealTypes().Create(ctx, mt)
	require.NoError(t, err)
	require.NotZero(t, createdMT.ID)
	defer client.MealTypes().Delete(ctx, createdMT.ID)

	// Create 7 meal plans
	createdIDs := make([]int, 0, 7)
	base := time.Now().Add(24 * time.Hour).UTC()
	for i := 0; i < 7; i++ {
		from := base.Add(time.Duration(i) * time.Hour)
		to := base.Add(time.Duration(i+1) * time.Hour)
		mp := &mealplan.MealPlan{
			Title:      "Edge Plan",
			MealTypeID: createdMT.ID,
			Servings:   1,
			FromDate:   &from,
			ToDate:     &to,
			Note:       "test",
		}
		created, err := client.MealPlans().Create(ctx, mp)
		require.NoError(t, err)
		require.NotZero(t, created.ID)
		createdIDs = append(createdIDs, created.ID)
	}
	// Cleanup
	defer func() {
		for _, id := range createdIDs {
			_ = client.MealPlans().Delete(ctx, id)
		}
	}()

	// Page size 3, page 1
	opts := &mealplan.ListOptions{ListOptions: pagination.ListOptions{Page: 1, PageSize: 3}}
	page1, err := client.MealPlans().List(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, 7, page1.Count)
	require.Len(t, page1.Results, 3)
	require.True(t, page1.HasNext())

	// Page 3 (last partial)
	opts.Page = 3
	page3, err := client.MealPlans().List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, page3.Results, 1)
	require.False(t, page3.HasNext())

	// Page beyond last
	opts.Page = 10
	_, err = client.MealPlans().List(ctx, opts)
	// Tandoor returns 404 for invalid page
	require.Error(t, err)
}
