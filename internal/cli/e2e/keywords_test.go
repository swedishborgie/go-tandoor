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

func TestE2EKeywordsList(t *testing.T) {
	out, _, code, err := runCLI("keywords", "list")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2EKeywordsCreateUpdate(t *testing.T) {
	name := "e2e-keyword-" + t.Name()
	newName := name + "-updated"
	// create
	outCreate, errOut, code, err := runCLI("keywords", "create", name, "--label", name)
	t.Logf("create out=%s errOut=%s code=%d err=%v", outCreate, errOut, code, err)
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
	require.Equal(t, name, created.Name)
	// update
	outUpd, _, code, err := runCLI("keywords", "update", fmt.Sprintf("%d", created.ID), "--name", newName)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outUpd, newName)
	// verify
	outGet, _, code, err := runCLI("keywords", "get", fmt.Sprintf("%d", created.ID))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outGet, newName)
	// cleanup
	_, _, _, _ = runCLI("keywords", "delete", fmt.Sprintf("%d", created.ID))
}
