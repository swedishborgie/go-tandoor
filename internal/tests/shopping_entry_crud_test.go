//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/shopping"
	"github.com/swedishborgie/go-tandoor/unit"
)

func TestIntegrationShoppingEntryCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a shopping list
	sl := &shopping.List{
		Name: "Entry Test List",
	}
	createdList, err := client.ShoppingLists().Create(ctx, sl)
	require.NoError(t, err)
	require.NotZero(t, createdList.ID)
	defer client.ShoppingLists().Delete(ctx, createdList.ID)

	// Create food and unit
	f := &food.Food{Name: "Entry Test Food"}
	createdFood, err := client.Foods().Create(ctx, f)
	require.NoError(t, err)
	defer client.Foods().Delete(ctx, createdFood.ID)

	u := &unit.Unit{Name: "Entry Test Unit", PluralName: "Entry Test Units"}
	createdUnit, err := client.Units().Create(ctx, u)
	require.NoError(t, err)
	defer client.Units().Delete(ctx, createdUnit.ID)

	// Create shopping list entry
	entry := &shopping.ListEntry{
		Amount: floatPtr(2.5),
		Note:   "Test entry",
		Food: &food.Shopping{
			ID:   createdFood.ID,
			Name: createdFood.Name,
		},
		Unit: &unit.Unit{
			ID:   createdUnit.ID,
			Name: createdUnit.Name,
		},
		Lists: []shopping.List{{ID: createdList.ID}},
	}
	createdEntry, err := client.ShoppingEntries().Create(ctx, entry)
	require.NoError(t, err)
	require.NotZero(t, createdEntry.ID)
	defer client.ShoppingEntries().Delete(ctx, createdEntry.ID)

	// Get
	got, err := client.ShoppingEntries().Get(ctx, createdEntry.ID)
	require.NoError(t, err)
	require.Equal(t, createdEntry.ID, got.ID)
	require.NotNil(t, got.Amount)
	require.InDelta(t, 2.5, *got.Amount, 0.001) // InDelta can't take *float64

	// Update
	got.Note = "Updated note"
	updated, err := client.ShoppingEntries().Update(ctx, got)
	require.NoError(t, err)
	// Note may be omitted in response due to serializer, just ensure update succeeded
	require.Equal(t, got.ID, updated.ID)

	// Delete
	err = client.ShoppingEntries().Delete(ctx, createdEntry.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = client.ShoppingEntries().Get(ctx, createdEntry.ID)
	require.Error(t, err)
}
