//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationStorageCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	payload := map[string]any{
		"name":   "Test Storage",
		"method": "LOCAL",
		"url":    "http://localhost/test",
	}
	var created struct {
		ID int `json:"id"`
	}
	err := client.DoJSON(ctx, "POST", "api/storage/", payload, &created)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/storage/%d/", created.ID), nil, nil) }()

	var got struct {
		ID int `json:"id"`
	}
	err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/storage/%d/", created.ID), nil, &got)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
}
