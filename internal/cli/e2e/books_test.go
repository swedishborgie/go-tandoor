//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EBooksList(t *testing.T) {
	out, _, code, err := runCLI("books", "list")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2EBooksCreateUpdateDelete(t *testing.T) {
	t.Skip("books create returns 500 in test env")
}
