//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/recipe"
)

func TestIntegrationPaginationIterator(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()

	const total = 23
	ids := make([]int, 0, total)
	for i := 0; i < total; i++ {
		r := &recipe.Recipe{
			Name:     "Iterator Test Recipe",
			Servings: 1,
			Steps:    []recipe.Step{},
		}
		created, err := client.Recipes().Create(ctx, r)
		require.NoError(t, err)
		ids = append(ids, created.ID)
	}

	pageSize := 10
	page := 1
	var count int
	iterated := 0

	// Use iterator pattern
	for {
		opts := &recipe.ListOptions{
			ListOptions: pagination.ListOptions{Page: page, PageSize: pageSize},
		}
		p, err := client.Recipes().List(ctx, opts)
		require.NoError(t, err)
		if page == 1 {
			count = p.Count
		}
		iterated += len(p.Results)
		if !p.HasNext() {
			break
		}
		page++
	}

	require.Equal(t, total, count)
	require.Equal(t, total, iterated)

	// Cleanup
	for _, id := range ids {
		_ = client.Recipes().Delete(ctx, id)
	}
}
