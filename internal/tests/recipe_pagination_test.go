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

func TestIntegrationRecipePagination(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()

	// Create 15 recipes
	const total = 15
	ids := make([]int, 0, total)
	for i := 0; i < total; i++ {
		r := &recipe.Recipe{
			Name:     "Pagination Test Recipe",
			Servings: 1,
			Steps:    []recipe.Step{},
		}
		created, err := client.Recipes().Create(ctx, r)
		require.NoError(t, err)
		require.NotNil(t, created)
		ids = append(ids, created.ID)
	}

	// List with page size 10
	opts := &recipe.ListOptions{
		ListOptions: pagination.ListOptions{Page: 1, PageSize: 10},
	}
	page1, err := client.Recipes().List(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, page1)
	require.Equal(t, total, page1.Count)
	require.Len(t, page1.Results, 10)
	require.True(t, page1.HasNext())

	// Page 2
	opts.Page = 2
	page2, err := client.Recipes().List(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, page2)
	require.Len(t, page2.Results, 5)
	require.False(t, page2.HasNext())

	// Cleanup
	for _, id := range ids {
		_ = client.Recipes().Delete(ctx, id)
	}
}
