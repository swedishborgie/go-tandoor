//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/food"
)

func TestIntegrationInventoryCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a food
	f := &food.Food{Name: "Inventory Test Food"}
	createdFood, err := client.Foods().Create(ctx, f)
	require.NoError(t, err)
	require.NotZero(t, createdFood.ID)
	defer client.Foods().Delete(ctx, createdFood.ID)

	// Find existing inventory location created in SetupTestData
	var locList struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	err = client.DoJSON(ctx, "GET", "api/inventory-location/", nil, &locList)
	require.NoError(t, err)
	require.Greater(t, len(locList.Results), 0)
	locID := locList.Results[0].ID

	// Create a unit
	unitPayload := map[string]any{
		"name": "Test Unit",
	}
	var unit struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "POST", "api/unit/", unitPayload, &unit)
	require.NoError(t, err)
	require.NotZero(t, unit.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/unit/%d/", unit.ID), nil, nil) }()
	unitID := unit.ID

	// Create inventory entry
	entryPayload := map[string]any{
		"food":               createdFood.ID,
		"inventory_location": locID,
		"unit":               unitID,
		"amount":             5.0,
		"code":               "TEST-001",
	}
	var entry struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "POST", "api/inventory-entry/", entryPayload, &entry)
	require.NoError(t, err)
	require.NotZero(t, entry.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/inventory-entry/%d/", entry.ID), nil, nil) }()

	// Get entry
	var got struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/inventory-entry/%d/", entry.ID), nil, &got)
	require.NoError(t, err)
	require.Equal(t, entry.ID, got.ID)
}
