//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EUnitsList(t *testing.T) {
	out, _, code, err := runCLI("units", "list")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2EUnitsCreateUpdate(t *testing.T) {
	name := "e2e-unit-" + t.Name()
	newName := name + "-updated"
	// create
	outCreate, _, code, err := runCLI("units", "create", name)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	idx := strings.Index(outCreate, "{")
	require.Greater(t, idx, -1)
	var created struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(outCreate[idx:]), &created))
	require.Greater(t, created.ID, 0)
	// update
	outUpd, _, code, err := runCLI("units", "update", fmt.Sprintf("%d", created.ID), "--name", newName)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outUpd, newName)
	// cleanup
	_, _, _, _ = runCLI("units", "delete", fmt.Sprintf("%d", created.ID))
}
