//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/mealplan"
)

func TestIntegrationMealTypeCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create
	mt := &mealplan.MealType{
		Name:  "Integration Test MealType",
		Order: 99,
		Time:  "12:00:00",
		Color: "#123456",
	}
	created, err := client.MealTypes().Create(ctx, mt)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	defer client.MealTypes().Delete(ctx, created.ID)

	// Get
	got, err := client.MealTypes().Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, mt.Name, got.Name)

	// Update
	got.Name = "Integration Test MealType Updated"
	updated, err := client.MealTypes().Update(ctx, got)
	require.NoError(t, err)
	require.Equal(t, "Integration Test MealType Updated", updated.Name)

	// Delete
	err = client.MealTypes().Delete(ctx, updated.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = client.MealTypes().Get(ctx, updated.ID)
	require.Error(t, err)
}
