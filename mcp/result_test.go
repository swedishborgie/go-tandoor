package mcp

import (
	"context"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonResultMarshalError(t *testing.T) {
	res := jsonResult(make(chan int)) // channels do not marshal
	require.True(t, res.IsError)
}

func TestJsonResultJQ(t *testing.T) {
	ctx := context.Background()

	// No jq arg: plain JSON.
	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	res, err := jsonResultJQ(ctx, req, map[string]any{"a": 1})
	require.NoError(t, err)
	assert.False(t, res.IsError)

	// Valid jq.
	req = mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]any{"jq": ".a"}
	res, err = jsonResultJQ(ctx, req, map[string]any{"a": 1})
	require.NoError(t, err)
	assert.False(t, res.IsError)

	// Invalid jq filter.
	req = mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]any{"jq": ".["}
	res, err = jsonResultJQ(ctx, req, map[string]any{"a": 1})
	require.NoError(t, err)
	assert.True(t, res.IsError)
}

func TestApplyJQ(t *testing.T) {
	ctx := context.Background()

	out, err := applyJQ(ctx, ".x", map[string]any{"x": 42})
	require.NoError(t, err)
	assert.Contains(t, out, "42")

	// Runtime jq error.
	_, err = applyJQ(ctx, `error("boom")`, 1)
	require.Error(t, err)

	// halt (nil value) stops iteration cleanly and yields "null".
	out, err = applyJQ(ctx, "halt", 1)
	require.NoError(t, err)
	assert.Equal(t, "null", out)

	// Empty result set yields "null".
	out, err = applyJQ(ctx, "empty", 1)
	require.NoError(t, err)
	assert.Equal(t, "null", out)

	// Malformed filter.
	_, err = applyJQ(ctx, ".[", 1)
	require.Error(t, err)
}

func TestBoolArg(t *testing.T) {
	// nil arguments.
	var req mcpgo.CallToolRequest
	v, ok := boolArg(req, "flag")
	assert.False(t, v)
	assert.False(t, ok)

	// Missing key.
	req = mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	_, ok = boolArg(req, "flag")
	assert.False(t, ok)

	// Non-bool value.
	req.Params.Arguments = map[string]any{"flag": "yes"}
	v, ok = boolArg(req, "flag")
	assert.False(t, v)
	assert.True(t, ok)

	// Explicit false.
	req.Params.Arguments = map[string]any{"flag": false}
	v, ok = boolArg(req, "flag")
	assert.False(t, v)
	assert.True(t, ok)

	// Explicit true.
	req.Params.Arguments = map[string]any{"flag": true}
	v, ok = boolArg(req, "flag")
	assert.True(t, v)
	assert.True(t, ok)
}

func TestErrResultPlain(t *testing.T) {
	res := errResult(assert.AnError)
	require.True(t, res.IsError)
}
