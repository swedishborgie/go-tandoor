//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EMealPlansList(t *testing.T) {
	out, _, code, err := runCLI("mealplans", "list")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2EMealPlansCreateUpdateDelete(t *testing.T) {
	t.Skip("meal plans create still requires additional fields")
}
