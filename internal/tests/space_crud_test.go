//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationSpaceCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create Space
	payload := map[string]any{
		"name": "Test Space",
	}
	var created struct {
		ID int `json:"id"`
	}
	err := client.DoJSON(ctx, "POST", "api/space/", payload, &created)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/space/%d/", created.ID), nil, nil) }()

	// Get Space
	var got struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/space/%d/", created.ID), nil, &got)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Test Space", got.Name)

	// Update Space
	updatePayload := map[string]any{
		"name": "Updated Space",
	}
	var updated struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err = client.DoJSON(ctx, "PATCH", fmt.Sprintf("api/space/%d/", created.ID), updatePayload, &updated)
	require.NoError(t, err)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, "Updated Space", updated.Name)

	// List Spaces
	var list struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	err = client.DoJSON(ctx, "GET", "api/space/", nil, &list)
	require.NoError(t, err)
	found := false
	for _, r := range list.Results {
		if r.ID == created.ID {
			found = true
			break
		}
	}
	require.True(t, found)
}
