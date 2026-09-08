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

func TestE2ERecipesList(t *testing.T) {
	out, _, code, err := runCLI("recipes", "list", "--page-size", "5")
	require.Equal(t, 0, code, "CLI should exit 0")
	require.NoError(t, err)
	require.Contains(t, out, "results")
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2ERecipesCreateDryRun(t *testing.T) {
	name := "e2e-cli-recipe-dryrun-" + t.Name()
	out, _, code, err := runCLI("recipes", "create", "--name", name, "--dry-run")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, out, name)
	// ensure no recipe was created
	listOut, _, _, _ := runCLI("recipes", "list", "--search", name, "--page-size", "10")
	require.NotContains(t, listOut, name)
}

func TestE2ERecipesCreateAndGet(t *testing.T) {
	// List an existing recipe and get it via CLI
	outList, _, code, err := runCLI("recipes", "list", "--page-size", "1")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var list struct {
		Results []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(outList), &list))
	var id int
	var name string
	if len(list.Results) == 0 {
		name = "seed-recipe-" + t.Name()
		outCreate, _, code, err := runCLI("recipes", "create", "--name", name, "--servings", "2")
		require.Equal(t, 0, code)
		require.NoError(t, err)
		idx := strings.Index(outCreate, "{")
		require.Greater(t, idx, -1)
		var created struct {
			ID int `json:"id"`
		}
		require.NoError(t, json.Unmarshal([]byte(outCreate[idx:]), &created))
		id = created.ID
		defer func() { runCLI("recipes", "delete", fmt.Sprintf("%d", id)) }()
	} else {
		id = list.Results[0].ID
		name = list.Results[0].Name
	}

	outGet, _, code, err := runCLI("recipes", "get", fmt.Sprintf("%d", id))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outGet, name)
}

func TestE2ERecipesGetByID(t *testing.T) {
	outList, _, code, err := runCLI("recipes", "list", "--page-size", "1")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var list struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(outList), &list))
	var id int
	if len(list.Results) == 0 {
		name := "seed-get-" + t.Name()
		outCreate, _, code, err := runCLI("recipes", "create", "--name", name, "--servings", "2")
		require.Equal(t, 0, code)
		require.NoError(t, err)
		idx := strings.Index(outCreate, "{")
		require.Greater(t, idx, -1)
		var created struct {
			ID int `json:"id"`
		}
		require.NoError(t, json.Unmarshal([]byte(outCreate[idx:]), &created))
		id = created.ID
		defer func() { runCLI("recipes", "delete", fmt.Sprintf("%d", id)) }()
	} else {
		id = list.Results[0].ID
	}
	out, _, code, err := runCLI("recipes", "get", fmt.Sprintf("%d", id))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, out, fmt.Sprintf("%d", id))
}

func TestE2ERecipesUpdate(t *testing.T) {
	name := "e2e-update-" + t.Name()
	updatedName := name + "-updated"
	// create
	outCreate, _, code, err := runCLI("recipes", "create", "--name", name, "--servings", "2")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	idx := strings.Index(outCreate, "{")
	require.Greater(t, idx, -1)
	var created struct {
		ID int `json:"id"`
	}
	require.NoError(t, json.Unmarshal([]byte(outCreate[idx:]), &created))
	// update
	outUpd, _, code, err := runCLI("recipes", "update", fmt.Sprintf("%d", created.ID), "--name", updatedName)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outUpd, updatedName)
	// get and verify
	outGet, _, code, err := runCLI("recipes", "get", fmt.Sprintf("%d", created.ID))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outGet, updatedName)
	// cleanup
	_, _, _, _ = runCLI("recipes", "delete", fmt.Sprintf("%d", created.ID))
}

func TestE2ERecipesCreateAndCleanup(t *testing.T) {
	// t.Skip("recipes create 500 in test env – skip for now")
	name := "test"
	// create real recipe
	outCreate, _, code, err := runCLI("recipes", "create", "--name", name, "--servings", "2")
	t.Logf("create out=%s code=%d err=%v", outCreate, code, err)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outCreate, name)
	// extract JSON part after first '{'
	idx := strings.Index(outCreate, "{")
	require.Greater(t, idx, -1)
	jsonPart := outCreate[idx:]
	var created struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal([]byte(jsonPart), &created))
	require.Greater(t, created.ID, 0)
	// get it back
	outGet, _, code, err := runCLI("recipes", "get", fmt.Sprintf("%d", created.ID))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outGet, name)
	// delete it
	_, _, code, err = runCLI("recipes", "delete", fmt.Sprintf("%d", created.ID))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	// verify gone
	outList, _, _, _ := runCLI("recipes", "list", "--search", name, "--page-size", "5")
	require.NotContains(t, outList, name)
}
