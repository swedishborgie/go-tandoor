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

func TestIntegrationPropertyCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create a food to attach property to
	f := &food.Food{Name: "Property Test Food"}
	createdFood, err := client.Foods().Create(ctx, f)
	require.NoError(t, err)
	require.NotZero(t, createdFood.ID)
	defer client.Foods().Delete(ctx, createdFood.ID)

	// Create a property type via raw payload
	ptCreate := map[string]any{
		"name":        "Test Property Type",
		"unit":        "g",
		"description": "Test",
		"order":       1,
	}
	var pt struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "POST", "api/property-type/", ptCreate, &pt)
	require.NoError(t, err)
	require.NotZero(t, pt.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/property-type/%d/", pt.ID), nil, nil) }()

	// Create property via raw payload to avoid serializer nesting issues
	propPayload := map[string]any{
		"food":            createdFood.ID,
		"property_type":   pt.ID,
		"property_amount": 10.5,
	}
	var createdProp struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "POST", "api/property/", propPayload, &createdProp)
	require.NoError(t, err)
	require.NotZero(t, createdProp.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/property/%d/", createdProp.ID), nil, nil) }()

	var got struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/property/%d/", createdProp.ID), nil, &got)
	require.NoError(t, err)
	require.Equal(t, createdProp.ID, got.ID)
}
