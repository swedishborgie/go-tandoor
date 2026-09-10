package recipe

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverview_MarshalJSON_IDOnly(t *testing.T) {
	b, err := json.Marshal(&Overview{ID: 7})
	require.NoError(t, err)
	assert.JSONEq(t, `7`, string(b))
}

func TestOverview_MarshalJSON_Full(t *testing.T) {
	b, err := json.Marshal(&Overview{ID: 7, Name: "Pasta", Servings: 4})
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.EqualValues(t, 7, m["id"])
	assert.Equal(t, "Pasta", m["name"])
	assert.EqualValues(t, 4, m["servings"])
}

func TestOverview_UnmarshalJSON_BareID(t *testing.T) {
	var o Overview
	require.NoError(t, json.Unmarshal([]byte(`7`), &o))
	assert.Equal(t, 7, o.ID)
	assert.Empty(t, o.Name)
}

func TestOverview_UnmarshalJSON_Object(t *testing.T) {
	var o Overview
	require.NoError(t, json.Unmarshal([]byte(`{"id":7,"name":"Pasta","servings":4}`), &o))
	assert.Equal(t, 7, o.ID)
	assert.Equal(t, "Pasta", o.Name)
	assert.Equal(t, 4, o.Servings)
}

func TestUser_MarshalJSON_IDOnly(t *testing.T) {
	b, err := json.Marshal(&User{ID: 3})
	require.NoError(t, err)
	assert.JSONEq(t, `3`, string(b))
}

func TestUser_MarshalJSON_Full(t *testing.T) {
	b, err := json.Marshal(&User{ID: 3, Username: "alice"})
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.EqualValues(t, 3, m["id"])
	assert.Equal(t, "alice", m["username"])
}

func TestUser_UnmarshalJSON_BareID(t *testing.T) {
	var u User
	require.NoError(t, json.Unmarshal([]byte(`3`), &u))
	assert.Equal(t, 3, u.ID)
	assert.Empty(t, u.Username)
}

func TestUser_UnmarshalJSON_Object(t *testing.T) {
	var u User
	require.NoError(t, json.Unmarshal([]byte(`{"id":3,"username":"alice"}`), &u))
	assert.Equal(t, 3, u.ID)
	assert.Equal(t, "alice", u.Username)
}
