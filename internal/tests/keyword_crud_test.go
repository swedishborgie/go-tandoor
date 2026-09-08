//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/keyword"
)

func TestIntegrationKeywordCRUD(t *testing.T) {
	client := Client()
	require.NotNil(t, client)
	ctx := context.Background()

	kw := &keyword.Keyword{
		Name: "Integration Test Keyword",
	}
	created, err := client.Keywords().Create(ctx, kw)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	defer client.Keywords().Delete(ctx, created.ID)

	got, err := client.Keywords().Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, kw.Name, got.Name)

	got.Name = "Integration Test Keyword Updated"
	updated, err := client.Keywords().Update(ctx, got)
	require.NoError(t, err)
	require.Equal(t, "Integration Test Keyword Updated", updated.Name)

	err = client.Keywords().Delete(ctx, updated.ID)
	require.NoError(t, err)

	_, err = client.Keywords().Get(ctx, updated.ID)
	require.Error(t, err)
}
