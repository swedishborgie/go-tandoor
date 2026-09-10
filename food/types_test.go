package food

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShopping_MarshalJSON_IDOnly(t *testing.T) {
	b, err := json.Marshal(&Shopping{ID: 4})
	require.NoError(t, err)
	assert.JSONEq(t, `4`, string(b))
}

func TestShopping_MarshalJSON_Full(t *testing.T) {
	b, err := json.Marshal(&Shopping{ID: 4, Name: "My list"})
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "My list", m["name"])
}

func TestShopping_UnmarshalJSON_BareID(t *testing.T) {
	var s Shopping
	require.NoError(t, json.Unmarshal([]byte(`4`), &s))
	assert.Equal(t, 4, s.ID)
}

func TestShopping_UnmarshalJSON_Object(t *testing.T) {
	var s Shopping
	require.NoError(t, json.Unmarshal([]byte(`{"id":4,"name":"My list"}`), &s))
	assert.Equal(t, 4, s.ID)
	assert.Equal(t, "My list", s.Name)
}
