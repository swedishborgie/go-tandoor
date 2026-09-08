//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EFoodsList(t *testing.T) {
	out, _, code, err := runCLI("foods", "list", "--page-size", "5")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}

func TestE2EFoodsListAll(t *testing.T) {
	out, _, code, err := runCLI("foods", "list", "--all")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	// should be a slice, possibly empty but valid
	_ = data.Results
}

func TestE2EFoodsEnsureDryRun(t *testing.T) {
	name := "e2e-test-food-ensure"
	out, _, code, err := runCLI("foods", "ensure", "--name", name, "--dry-run")
	t.Logf("foods ensure dry-run out: %s code:%d err:%v", out, code, err)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, out, name)
	// ensure no food created
	listOut, _, _, _ := runCLI("foods", "list", "--name-exact", name)
	// list returns empty or not containing name
	require.NotContains(t, listOut, name)
}

func TestE2EFoodsEnsureCreate(t *testing.T) {
	name := "e2e-test-food-create-" + t.Name()
	out, _, code, err := runCLI("foods", "ensure", "--name", name, "--force-create")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var result struct {
		Summary struct {
			Created int `json:"created"`
		} `json:"summary"`
		Results map[string]struct {
			Status string `json:"status"`
			Food   struct {
				ID int `json:"id"`
			} `json:"food"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.GreaterOrEqual(t, result.Summary.Created, 0)
	res, ok := result.Results[name]
	require.True(t, ok)
	require.Contains(t, []string{"created", "found"}, res.Status)
	// verify get works
	getOut, _, code, err := runCLI("foods", "get", fmt.Sprintf("%d", res.Food.ID))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, getOut, name)
}

func TestE2EFoodsUpdate(t *testing.T) {
	name := "e2e-food-update-" + t.Name()
	newName := name + "-updated"
	// ensure food exists
	outEnsure, _, code, err := runCLI("foods", "ensure", "--name", name, "--force-create")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var result struct {
		Results map[string]struct {
			Food struct {
				ID int `json:"id"`
			} `json:"food"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(outEnsure), &result))
	foodID := result.Results[name].Food.ID
	require.Greater(t, foodID, 0)
	// update name
	outUpd, _, code, err := runCLI("foods", "update", fmt.Sprintf("%d", foodID), "--name", newName)
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outUpd, newName)
	// verify
	outGet, _, code, err := runCLI("foods", "get", fmt.Sprintf("%d", foodID))
	require.Equal(t, 0, code)
	require.NoError(t, err)
	require.Contains(t, outGet, newName)
}
