//go:build integration
// +build integration

package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// newMCPClientWithEnv starts the binary with extra environment variables.
func newMCPClientWithEnv(t *testing.T, extraEnv ...string) *client.Client {
	t.Helper()
	env := append(os.Environ(), "TANDOOR_BASE_URL="+baseURL, "TANDOOR_TOKEN="+token)
	env = append(env, extraEnv...)
	mc, err := client.NewStdioMCPClient(mcpBin, env, "mcp")
	require.NoError(t, err)
	t.Cleanup(func() { _ = mc.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initReq := mcpgo.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpgo.Implementation{Name: "tandoor-mcp-e2e", Version: "0.0.0"}
	_, err = mc.Initialize(ctx, initReq)
	require.NoError(t, err)
	return mc
}

func listToolNames(t *testing.T, c *client.Client) map[string]bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	res, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	require.NoError(t, err)
	names := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	return names
}

func TestE2EReadOnlyMode(t *testing.T) {
	c := newMCPClientWithEnv(t, "TANDOOR_MCP_READ_ONLY=1")
	names := listToolNames(t, c)
	require.True(t, names["recipe_list"], "read tools must stay in read-only mode")
	require.True(t, names["server_info"])
	require.False(t, names["recipe_create"], "write tools must be hidden in read-only mode")
	require.False(t, names["keyword_delete"])

	// server_info reports the mode
	info := callTool(t, c, "server_info", nil)
	require.Equal(t, true, info["read_only"])
}

func TestE2EToolFilter(t *testing.T) {
	c := newMCPClientWithEnv(t, "TANDOOR_MCP_TOOLS=recipe_*,+server_info")
	names := listToolNames(t, c)
	require.True(t, names["recipe_list"])
	require.True(t, names["recipe_create"])
	require.True(t, names["server_info"])
	require.False(t, names["food_list"], "non-matching tools must be filtered out")
}
