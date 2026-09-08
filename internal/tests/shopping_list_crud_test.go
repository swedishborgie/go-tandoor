//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/shopping"
)

func TestIntegrationShoppingListCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()

	list := &shopping.List{
		Name:        "Integration Test List",
		Description: "Test description",
		Color:       "#ff0000",
	}
	created, err := client.ShoppingLists().Create(ctx, list)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Greater(t, created.ID, 0)

	got, err := client.ShoppingLists().Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	got.Name = "Integration Test List Updated"
	updated, err := client.ShoppingLists().Update(ctx, got)
	require.NoError(t, err)
	require.Equal(t, "Integration Test List Updated", updated.Name)

	err = client.ShoppingLists().Delete(ctx, updated.ID)
	require.NoError(t, err)

	_, err = client.ShoppingLists().Get(ctx, updated.ID)
	require.Error(t, err)
}
