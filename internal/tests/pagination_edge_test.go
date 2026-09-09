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

	// Baseline: other tests in this package may leave foods behind, so assert
	// relative counts rather than an absolute total.
	var baseline struct {
		Count int `json:"count"`
	}
	err := client.DoJSON(ctx, "GET", "api/food/?page_size=1", nil, &baseline)
	require.NoError(t, err)

	// Create 3 foods
	createdIDs := make(map[int]bool, 3)
	for i := 0; i < 3; i++ {
		f := map[string]any{"name": fmt.Sprintf("Edge Food %d", i)}
		var created struct {
			ID int `json:"id"`
		}
		err := client.DoJSON(ctx, "POST", "api/food/", f, &created)
		require.NoError(t, err)
		createdIDs[created.ID] = true
		defer func() { _ = client.DoJSON(ctx, "DELETE", fmt.Sprintf("api/food/%d/", created.ID), nil, nil) }()
	}

	// List with page size 2, page 1
	type pageResp struct {
		Count   int `json:"count"`
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
		Next *string `json:"next"`
	}
	var page1 pageResp
	err = client.DoJSON(ctx, "GET", "api/food/?page=1&page_size=2", nil, &page1)
	require.NoError(t, err)
	require.Equal(t, baseline.Count+3, page1.Count)
	require.Len(t, page1.Results, 2)

	// Walk all pages and verify the total adds up (with a non-zero baseline
	// the final page size depends on baseline % 2).
	seen := map[int]bool{}
	for _, r := range page1.Results {
		seen[r.ID] = true
	}
	for page := 2; ; page++ {
		var pr pageResp
		err = client.DoJSON(ctx, "GET", fmt.Sprintf("api/food/?page=%d&page_size=2", page), nil, &pr)
		require.NoError(t, err)
		for _, r := range pr.Results {
			seen[r.ID] = true
		}
		if pr.Next == nil || *pr.Next == "" {
			break
		}
		if page > 1000 {
			t.Fatal("pagination did not terminate")
		}
	}
	require.Equal(t, baseline.Count+3, len(seen))
	for id := range createdIDs {
		require.True(t, seen[id], "created food %d not found in pages", id)
	}

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
