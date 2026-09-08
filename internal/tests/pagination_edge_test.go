//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationPaginationEdge(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create 3 foods
	for i := 0; i < 3; i++ {
		f := map[string]any{"name": fmt.Sprintf("Edge Food %d", i)}
		var created struct {
			ID int `json:"id"`
		}
		err := client.DoJSON(ctx, "POST", "api/food/", f, &created)
		require.NoError(t, err)
		defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/food/%d/", created.ID), nil, nil) }()
	}

	// List with page size 2, page 1
	var page1 struct {
		Count   int `json:"count"`
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
		Next     *string `json:"next"`
		Previous *string `json:"previous"`
	}
	err := client.DoJSON(ctx, "GET", "api/food/?page=1&page_size=2", nil, &page1)
	require.NoError(t, err)
	require.Equal(t, 3, page1.Count)
	require.Len(t, page1.Results, 2)

	// Page 2
	var page2 struct {
		Count   int `json:"count"`
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	err = client.DoJSON(ctx, "GET", "api/food/?page=2&page_size=2", nil, &page2)
	require.NoError(t, err)
	require.Len(t, page2.Results, 1)

	// Page beyond last – API returns 404 Invalid page
	var page99 struct {
		Count   int `json:"count"`
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	err = client.DoJSON(ctx, "GET", "api/food/?page=99&page_size=2", nil, &page99)
	// Expect error due to 404
	require.Error(t, err)
}
