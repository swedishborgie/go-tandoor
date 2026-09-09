//go:build integration
// +build integration

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

var (
	baseURL = "http://localhost:8081"
	token   string
	mcpBin  string
)

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		os.Exit(m.Run())
	}
	if err := WriteEnvFile(); err != nil {
		panic(fmt.Errorf("WriteEnvFile: %w", err))
	}
	if err := Start(); err != nil {
		panic(fmt.Errorf("Start: %w", err))
	}
	if err := waitReady(); err != nil {
		panic(fmt.Errorf("waitReady: %w", err))
	}
	if err := CreateSuperUser(); err != nil {
		panic(fmt.Errorf("CreateSuperUser: %w", err))
	}
	if err := SetupTestData(); err != nil {
		panic(fmt.Errorf("SetupTestData: %w", err))
	}
	var err error
	token, err = GetAPIToken(baseURL)
	if err != nil {
		panic(fmt.Errorf("GetAPIToken: %w", err))
	}
	mcpBin, err = buildMCP()
	if err != nil {
		panic(fmt.Errorf("build MCP: %w", err))
	}
	code := m.Run()
	_ = Stop()
	os.Exit(code)
}

func waitReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 2 * time.Second}
	url := baseURL + "/"
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for tandoor")
		default:
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
}

// buildMCP compiles the tandoor-mcp binary once per suite.
func buildMCP() (string, error) {
	tmp := filepath.Join(os.TempDir(), "tandoor-mcp-e2e")
	cmd := exec.Command("go", "build", "-o", tmp, "./cmd/tandoor-mcp")
	cmd.Dir = repoRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("go build failed: %v\n%s", err, out)
	}
	return tmp, nil
}

// repoRoot returns the repository root, derived from this file's location
// (internal/mcp/e2e is three levels below the root).
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

// newMCPClient starts the tandoor-mcp binary as a subprocess and returns an
// initialized in-process stdio client.
func newMCPClient(t *testing.T) *client.Client {
	t.Helper()
	env := append(os.Environ(), "TANDOOR_BASE_URL="+baseURL, "TANDOOR_TOKEN="+token)
	mc, err := client.NewStdioMCPClient(mcpBin, env)
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

// callTool invokes a tool and returns the parsed JSON object result,
// failing the test if the tool reports an error.
func callTool(t *testing.T, c *client.Client, name string, args map[string]any) map[string]any {
	t.Helper()
	req := mcpgo.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	res, err := c.CallTool(ctx, req)
	require.NoError(t, err)
	require.Falsef(t, res.IsError, "tool %s returned error: %v", name, res.Content)
	require.NotEmpty(t, res.Content)
	text, ok := res.Content[0].(mcpgo.TextContent)
	require.True(t, ok, "expected text content, got %T", res.Content[0])
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(text.Text), &out))
	return out
}

// callToolRaw invokes a tool and returns the raw text result (for tools that
// may legitimately return an error payload the caller wants to inspect).
func callToolRaw(t *testing.T, c *client.Client, name string, args map[string]any) (string, bool) {
	t.Helper()
	req := mcpgo.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	res, err := c.CallTool(ctx, req)
	require.NoError(t, err)
	var text string
	if len(res.Content) > 0 {
		if tc, ok := res.Content[0].(mcpgo.TextContent); ok {
			text = tc.Text
		}
	}
	return text, res.IsError
}
