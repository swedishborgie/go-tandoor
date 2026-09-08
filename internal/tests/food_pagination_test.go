//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestIntegrationFoodPagination(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()

	const total = 13
	ids := make([]int, 0, total)
	for i := 0; i < total; i++ {
		f := &food.Food{
			Name: "Pagination Food " + string(rune('A'+i)),
		}
		created, err := client.Foods().Create(ctx, f)
		require.NoError(t, err)
		ids = append(ids, created.ID)
	}

	opts := &food.ListOptions{
		ListOptions: pagination.ListOptions{Page: 1, PageSize: 5},
	}
	page1, err := client.Foods().List(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, total, page1.Count)
	require.Len(t, page1.Results, 5)
	require.True(t, page1.HasNext())

	opts.Page = 3
	page3, err := client.Foods().List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, page3.Results, 3) // 13 total, 5+5+3

	// Cleanup
	for _, id := range ids {
		_ = client.Foods().Delete(ctx, id)
	}
}
