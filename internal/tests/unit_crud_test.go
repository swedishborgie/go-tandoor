//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/unit"
)

func TestIntegrationUnitCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)

	ctx := context.Background()

	u := &unit.Unit{
		Name:       "Integration Test Unit",
		PluralName: "Integration Test Units",
	}
	created, err := client.Units().Create(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "Integration Test Unit", created.Name)
	require.Greater(t, created.ID, 0)

	got, err := client.Units().Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	got.Name = "Integration Test Unit Updated"
	updated, err := client.Units().Update(ctx, got)
	require.NoError(t, err)
	require.Equal(t, "Integration Test Unit Updated", updated.Name)

	err = client.Units().Delete(ctx, updated.ID)
	require.NoError(t, err)

	_, err = client.Units().Get(ctx, updated.ID)
	require.Error(t, err)
}
