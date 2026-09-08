//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EPropertiesList(t *testing.T) {
	out, _, code, err := runCLI("properties", "list")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2EPropertiesTypesListAll(t *testing.T) {
	out, _, code, err := runCLI("properties", "types", "list", "--all")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.NotEmpty(t, data)
	require.Contains(t, data[0], "id")
	require.Contains(t, data[0], "name")
}

func TestE2EPropertiesTypesGet(t *testing.T) {
	listOut, _, code, err := runCLI("properties", "types", "list", "--all")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var types []map[string]any
	require.NoError(t, json.Unmarshal([]byte(listOut), &types))
	require.NotEmpty(t, types)
	id := fmt.Sprintf("%v", types[0]["id"])

	out, _, code, err := runCLI("properties", "types", "get", id)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Equal(t, types[0]["name"], data["name"])
}
