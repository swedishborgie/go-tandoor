package idref

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// plain is a minimal reference type using the standard two-method glue.
type plain struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
	Note string `json:"note,omitempty"`
}

func (p *plain) MarshalJSON() ([]byte, error) {
	type raw plain
	return MarshalJSON(p.ID, p.Name, raw(*p))
}

func (p *plain) UnmarshalJSON(data []byte) error {
	type raw plain
	var r raw
	if err := UnmarshalJSON(data, &r.ID, &r); err != nil {
		return err
	}
	*p = plain(r)
	return nil
}

func TestMarshalBareInt(t *testing.T) {
	p := plain{ID: 5}
	b, err := json.Marshal(&p)
	require.NoError(t, err)
	assert.Equal(t, `5`, string(b))
}

func TestMarshalFullObject(t *testing.T) {
	p := plain{ID: 5, Name: "gram"}
	b, err := json.Marshal(&p)
	require.NoError(t, err)
	assert.Equal(t, `{"id":5,"name":"gram"}`, string(b))
}

func TestMarshalFullObjectNoID(t *testing.T) {
	p := plain{Name: "new-unit"}
	b, err := json.Marshal(&p)
	require.NoError(t, err)
	assert.Equal(t, `{"name":"new-unit"}`, string(b))
}

func TestMarshalExtras(t *testing.T) {
	p := plain{ID: 3, Name: "x", Note: "d"}
	b, err := json.Marshal(&p)
	require.NoError(t, err)
	assert.Equal(t, `{"id":3,"name":"x","note":"d"}`, string(b))
}

func TestUnmarshalBareInt(t *testing.T) {
	var p plain
	require.NoError(t, json.Unmarshal([]byte(`12`), &p))
	assert.Equal(t, 12, p.ID)
	assert.Empty(t, p.Name)
	assert.Empty(t, p.Note)
}

func TestUnmarshalObject(t *testing.T) {
	var p plain
	require.NoError(t, json.Unmarshal([]byte(`{"id":4,"name":"cup","note":"n"}`), &p))
	assert.Equal(t, 4, p.ID)
	assert.Equal(t, "cup", p.Name)
	assert.Equal(t, "n", p.Note)
}

func TestUnmarshalNull(t *testing.T) {
	p := plain{ID: 9, Name: "old"}
	require.NoError(t, json.Unmarshal([]byte(`null`), &p))
	assert.Equal(t, 0, p.ID)
	assert.Empty(t, p.Name)
}

func TestRoundTripInContainer(t *testing.T) {
	type holder struct {
		Unit *plain  `json:"unit"`
		Ids  []plain `json:"ids"`
	}
	in := &holder{Unit: &plain{ID: 3}, Ids: []plain{{ID: 1}, {ID: 2}}}
	b, err := json.Marshal(in)
	require.NoError(t, err)
	assert.Equal(t, `{"unit":3,"ids":[1,2]}`, string(b))

	var out holder
	require.NoError(t, json.Unmarshal([]byte(`{"unit":8,"ids":[1,{"id":2,"name":"g"}]}`), &out))
	assert.Equal(t, 8, out.Unit.ID)
	assert.Equal(t, 2, out.Ids[1].ID)
	assert.Equal(t, "g", out.Ids[1].Name)
}

func TestMarshalJSON(t *testing.T) {
	b, err := MarshalJSON(9, "", "ignored")
	require.NoError(t, err)
	assert.Equal(t, `9`, string(b))

	b, err = MarshalJSON(9, "named", struct{ X int }{X: 1})
	require.NoError(t, err)
	assert.Equal(t, `{"X":1}`, string(b))
}

func TestUnmarshalJSONBareIntLeavesVUntouched(t *testing.T) {
	v := struct{ X int }{X: 1}
	id := 0
	require.NoError(t, UnmarshalJSON([]byte(`7`), &id, &v))
	assert.Equal(t, 7, id)
	assert.Equal(t, 1, v.X)
}

func TestAsID(t *testing.T) {
	id, ok := AsID([]byte(`7`))
	assert.True(t, ok)
	assert.Equal(t, 7, id)
	_, ok = AsID([]byte(`{"id":7}`))
	assert.False(t, ok)
	_, ok = AsID([]byte(`"7"`))
	assert.False(t, ok)
}
