// Package step contains types for Tandoor Step resources.
package step

import "encoding/json"

// Step is a cooking step within a recipe, as returned by the standalone
// /api/step/ endpoint.
type Step struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Instruction is the step body text.
	Instruction string `json:"instruction,omitempty"`
	// Ingredients are the IDs of the ingredients attached to this step. The
	// API requires the key on create and update (an empty list is valid), so
	// callers must set an empty slice rather than leave it nil for those
	// writes. Nil omits the key, which is the correct behavior for partial
	// updates. The API returns full ingredient objects; UnmarshalJSON
	// reduces them to IDs.
	Ingredients          []int `json:"ingredients,omitempty"`
	Time                 int   `json:"time,omitempty"`
	Order                int   `json:"order,omitempty"`
	ShowAsHeader         bool  `json:"show_as_header,omitempty"`
	ShowIngredientsTable bool  `json:"show_ingredients_table,omitempty"`
}

// MarshalJSON emits the ingredients key whenever Ingredients is non-nil
// (including an empty slice): the API requires the key on create/update, and
// omitempty alone cannot distinguish "not set" (nil, correct for PATCH) from
// "set to empty" (nil-safe []int{}). Field order in the output is not
// significant for the API.
func (s Step) MarshalJSON() ([]byte, error) {
	type plain Step
	p := plain(s)
	p.Ingredients = nil
	base, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	if s.Ingredients == nil {
		return base, nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}
	ids, err := json.Marshal(s.Ingredients)
	if err != nil {
		return nil, err
	}
	m["ingredients"] = ids
	return json.Marshal(m)
}

// UnmarshalJSON accepts ingredients as a list of IDs or a list of full
// ingredient objects (the latter is what the API returns).
func (s *Step) UnmarshalJSON(data []byte) error {
	type plain Step
	var p plain
	aux := struct {
		*plain
		Ingredients json.RawMessage `json:"ingredients"`
	}{plain: &p}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*s = Step(p)
	if len(aux.Ingredients) > 0 && string(aux.Ingredients) != "null" {
		if ids, err := ingredientIDs(aux.Ingredients); err == nil {
			s.Ingredients = ids
		}
	}
	return nil
}

func ingredientIDs(data []byte) ([]int, error) {
	var ids []int
	if err := json.Unmarshal(data, &ids); err == nil {
		return ids, nil
	}
	var objs []struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(data, &objs); err != nil {
		return nil, err
	}
	out := make([]int, 0, len(objs))
	for _, o := range objs {
		out = append(out, o.ID)
	}
	return out, nil
}
