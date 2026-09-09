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

func TestIntegrationShoppingEntryBulkUpdate(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a shopping list
	sl := &shopping.List{Name: "Bulk Test List"}
	createdList, err := client.ShoppingLists().Create(ctx, sl)
	require.NoError(t, err)
	require.NotZero(t, createdList.ID)
	defer client.ShoppingLists().Delete(ctx, createdList.ID)

	// Create food and unit
	f := &food.Food{Name: "Bulk Test Food"}
	createdFood, err := client.Foods().Create(ctx, f)
	require.NoError(t, err)
	defer client.Foods().Delete(ctx, createdFood.ID)

	u := &unit.Unit{Name: "Bulk Test Unit", PluralName: "Bulk Test Units"}
	createdUnit, err := client.Units().Create(ctx, u)
	require.NoError(t, err)
	defer client.Units().Delete(ctx, createdUnit.ID)

	// Create three entries
	entryIDs := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		entry := &shopping.ListEntry{
			Amount: floatPtr(float64(i + 1)),
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
		entryIDs = append(entryIDs, createdEntry.ID)
		// Cleanup
		defer client.ShoppingEntries().Delete(ctx, createdEntry.ID)
	}

	// Bulk update checked state
	checked := true
	bulk := &shopping.ListEntryBulk{
		IDs:     entryIDs,
		Checked: &checked,
	}
	updated, err := client.ShoppingEntries().BulkUpdate(ctx, bulk)
	require.NoError(t, err)
	require.NotNil(t, updated)

	// Verify each entry is checked
	for _, id := range entryIDs {
		got, err := client.ShoppingEntries().Get(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, got.Checked)
		require.True(t, *got.Checked)
	}
}
