package auditfood

import (
	"testing"
)

func TestJaccardSimilarity(t *testing.T) {
	cases := []struct {
		a, b string
		want float64
	}{
		{"chicken", "chicken", 1.0},
		{"Diced Chicken", "Chicken", 0.5}, // {diced,chicken} vs {chicken}: inter 1, union 2
		{"chicken", "beef", 0.0},
		{"", "", 1.0},
		{"", "chicken", 0.0},
	}
	for _, tc := range cases {
		got := JaccardSimilarity(tc.a, tc.b)
		if got < tc.want-1e-9 || got > tc.want+1e-9 {
			t.Errorf("JaccardSimilarity(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestWordSet(t *testing.T) {
	set := WordSet("Diced CHICKEN")
	if len(set) != 2 {
		t.Fatalf("len(set) = %d, want 2", len(set))
	}
	for _, w := range []string{"diced", "chicken"} {
		if _, ok := set[w]; !ok {
			t.Errorf("missing %q in set", w)
		}
	}
}

func TestUnitIDFromRef(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want int
	}{
		{"nil", nil, 0},
		{"int", 17, 17},
		{"float64", float64(42), 42},
		{"object", map[string]any{"id": float64(7), "name": "gram"}, 7},
		{"object no id", map[string]any{"name": "gram"}, 0},
		{"string", "gram", 0},
	}
	for _, tc := range cases {
		if got := unitIDFromRef(tc.v); got != tc.want {
			t.Errorf("%s: unitIDFromRef = %d, want %d", tc.name, got, tc.want)
		}
	}
}
