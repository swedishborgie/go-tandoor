//go:build integration
// +build integration

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2ECookLogsList(t *testing.T) {
	out, _, code, err := runCLI("cook-logs", "list")
	require.Equal(t, 0, code)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &data))
	require.Contains(t, data, "results")
}
