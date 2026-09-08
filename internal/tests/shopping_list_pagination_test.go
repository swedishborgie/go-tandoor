//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/shopping"
)

func TestIntegrationShoppingListPagination(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	// Create 13 shopping lists with unique names
	const total = 13
	createdIDs := make([]int, 0, total)
	for i := 0; i < total; i++ {
		list := &shopping.List{
			Name:        fmt.Sprintf("Pagination List %d", i),
			Description: "Pagination test",
		}
		created, err := client.ShoppingLists().Create(ctx, list)
		require.NoError(t, err)
		require.NotZero(t, created.ID)
		createdIDs = append(createdIDs, created.ID)
	}
	// Cleanup in defer
	defer func() {
		for _, id := range createdIDs {
			_ = client.ShoppingLists().Delete(ctx, id)
		}
	}()

	// List page 1 with page size 5
	opts := &shopping.ListListOptions{ListOptions: pagination.ListOptions{Page: 1, PageSize: 5}}
	page1, err := client.ShoppingLists().List(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, total, page1.Count)
	require.Len(t, page1.Results, 5)
	require.True(t, page1.HasNext())

	// List page 2
	opts.Page = 2
	page2, err := client.ShoppingLists().List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, page2.Results, 5)

	// List page 3
	opts.Page = 3
	page3, err := client.ShoppingLists().List(ctx, opts)
	require.NoError(t, err)
	require.Len(t, page3.Results, 3)

	// Verify no overlap
	ids := make(map[int]struct{})
	for _, l := range append(append(page1.Results, page2.Results...), page3.Results...) {
		ids[l.ID] = struct{}{}
	}
	require.Len(t, ids, total)
}
