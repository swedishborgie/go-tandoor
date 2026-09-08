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

func TestE2EShoppingListList(t *testing.T) {
	out, _, code, err := runCLI("shopping", "list", "list")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2EShoppingListCreate(t *testing.T) {
	name := "e2e-shop"
	out, _, code, err := runCLI("shopping", "create", name)
	t.Logf("shopping create out=%s code=%d err=%v", out, code, err)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, out, name)
	// cleanup
	var created struct {
		ID int `json:"id"`
	}
	_ = json.Unmarshal([]byte(out), &created)
	if created.ID != 0 {
		_, _, _, _ = runCLI("shopping", "delete", fmt.Sprintf("%d", created.ID))
	}
}

func TestE2EShoppingEntryAdd(t *testing.T) {
	listName := "e2e-shop-" + t.Name()
	foodName := "e2e-food-" + t.Name()
	// create list
	outList, errOut, code, err := runCLI("shopping", "create", listName)
	t.Logf("list create out=%s errOut=%s code=%d err=%v", outList, errOut, code, err)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	idx := strings.Index(outList, "{")
	require.Greater(t, idx, -1)
	var list struct {
		ID int `json:"id"`
	}
	require.NoError(t, json.Unmarshal([]byte(outList[idx:]), &list))
	require.Greater(t, list.ID, 0)
	// ensure food
	outFood, _, code, err := runCLI("foods", "ensure", "--name", foodName, "--force-create")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var foodRes struct {
		Results map[string]struct {
			Food struct {
				ID int `json:"id"`
			} `json:"food"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(outFood), &foodRes))
	foodID := foodRes.Results[foodName].Food.ID
	t.Logf("foodID=%d", foodID)
	require.Greater(t, foodID, 0)
	// add entry
	outEntry, errOut, code, err := runCLI("shopping", "add-entry", "--list-id", fmt.Sprintf("%d", list.ID), "--food-id", fmt.Sprintf("%d", foodID), "--amount", "1")
	t.Logf("add entry out=%s errOut=%s code=%d err=%v", outEntry, errOut, code, err)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outEntry, "id")
	// cleanup
	_, _, _, _ = runCLI("shopping", "delete", fmt.Sprintf("%d", list.ID))
}
