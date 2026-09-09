package step

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStepMarshalIngredients(t *testing.T) {
	// nil: key omitted (correct for PATCH)
	b, err := json.Marshal(Step{Name: "x", Instruction: "y"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "ingredients") {
		t.Fatalf("nil Ingredients should omit the key: %s", b)
	}

	// empty slice: key present with [] (required by the API on create/update)
	b, err = json.Marshal(Step{Name: "x", Instruction: "y", Ingredients: []int{}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"ingredients":[]`) {
		t.Fatalf("empty Ingredients should marshal as []: %s", b)
	}

	// populated slice: key present with IDs
	b, err = json.Marshal(Step{Instruction: "y", Ingredients: []int{1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"ingredients":[1,2]`) {
		t.Fatalf("Ingredients should marshal as [1,2]: %s", b)
	}
}

func TestStepUnmarshalIngredients(t *testing.T) {
	// list of full objects (what the API returns)
	var s Step
	if err := json.Unmarshal([]byte(`{"id":1,"instruction":"x","ingredients":[{"id":7,"note":"n"},{"id":9}]}`), &s); err != nil {
		t.Fatal(err)
	}
	if len(s.Ingredients) != 2 || s.Ingredients[0] != 7 || s.Ingredients[1] != 9 {
		t.Fatalf("expected [7 9], got %v", s.Ingredients)
	}

	// list of bare IDs
	s = Step{}
	if err := json.Unmarshal([]byte(`{"instruction":"x","ingredients":[3,4]}`), &s); err != nil {
		t.Fatal(err)
	}
	if len(s.Ingredients) != 2 || s.Ingredients[0] != 3 || s.Ingredients[1] != 4 {
		t.Fatalf("expected [3 4], got %v", s.Ingredients)
	}

	// missing key
	s = Step{}
	if err := json.Unmarshal([]byte(`{"instruction":"x"}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.Ingredients != nil {
		t.Fatalf("expected nil Ingredients, got %v", s.Ingredients)
	}
}
