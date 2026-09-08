//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2EFDCSearch(t *testing.T) {
	out, _, code, err := runCLI("fdc", "search", "--query", "tomato")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	// Output includes a header line printed to stderr, so just check JSON contains expected fields
	require.Contains(t, out, `"fdc_id": 123456`)
	require.Contains(t, out, `tomato mock food`)
}

func TestE2EFDCGet(t *testing.T) {
	out, _, code, err := runCLI("fdc", "get", "123456")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var resp struct {
		FDCID       int    `json:"fdc_id"`
		Description string `json:"description"`
		Nutrients   []struct {
			Number int `json:"number"`
		} `json:"nutrients"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &resp))
	require.Equal(t, 123456, resp.FDCID)
	require.Contains(t, resp.Description, "Mock Food")
	require.Len(t, resp.Nutrients, 4)
}
