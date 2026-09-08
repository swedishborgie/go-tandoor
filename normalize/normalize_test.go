package normalize

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantCanonical string
		wantConnector string
		wantNote      string
		wantAdj       string
	}{
		{
			name:          "quantity prefix with weight",
			input:         "(10 1/2 oz. each) beef broth",
			wantCanonical: "Beef Broth",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "",
		},
		{
			name:          "parenthetical mL can with product prep",
			input:         "(398mL) can diced tomatoes",
			wantCanonical: "Diced Tomatoes",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "diced",
		},
		{
			name:          "connector plus with quantity",
			input:         "+ 1 tablespoon all-purpose flour",
			wantCanonical: "All-Purpose Flour",
			wantConnector: "+",
			wantNote:      "",
			wantAdj:       "",
		},
		{
			name:          "connector minus with quantity",
			input:         "- 1 1/4 lb chicken thighs",
			wantCanonical: "Chicken Thighs",
			wantConnector: "-",
			wantNote:      "",
			wantAdj:       "",
		},
		{
			name:          "connector ampersand",
			input:         "& 1 orange sweet pepper",
			wantCanonical: "Orange Sweet Pepper",
			wantConnector: "&",
			wantNote:      "",
			wantAdj:       "",
		},
		{
			name:          "alternative clause",
			input:         "1 cup chicken broth or water",
			wantCanonical: "Chicken Broth",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "",
		},
		{
			name:          "leading fraction with user prep",
			input:         "½ medium red onion, sliced",
			wantCanonical: "Red Onion",
			wantConnector: "",
			wantNote:      "sliced",
			wantAdj:       "",
		},
		{
			name:          "boneless skinless chicken",
			input:         "1 ½ lbs boneless skinless chicken breast",
			wantCanonical: "Boneless Skinless Chicken Breast",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "boneless, skinless",
		},
		{
			name:          "shredded cheese",
			input:         "(2 cups) shredded mozzarella cheese",
			wantCanonical: "Shredded Mozzarella Cheese",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "shredded",
		},
		{
			name:          "frozen vegetables",
			input:         "frozen mixed vegetables",
			wantCanonical: "Frozen Mixed Vegetables",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "frozen",
		},
		{
			name:          "condensed soup",
			input:         "(10¾ oz.) condensed cream of mushroom soup",
			wantCanonical: "Condensed Cream Of Mushroom Soup",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "condensed",
		},
		{
			name:          "kosher salt with weight",
			input:         "(10 g) kosher salt",
			wantCanonical: "Kosher Salt",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "",
		},
		{
			name:          "black beans with can",
			input:         "(398mL) can black beans",
			wantCanonical: "Black Beans",
			wantConnector: "",
			wantNote:      "",
			wantAdj:       "",
		},
		{
			name:          "avocado oil with alternative",
			input:         "+ 1 teaspoon avocado oil or olive oil",
			wantCanonical: "Avocado Oil",
			wantConnector: "+",
			wantNote:      "",
			wantAdj:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			if result.Canonical != tt.wantCanonical {
				t.Errorf("Canonical = %q, want %q", result.Canonical, tt.wantCanonical)
			}
			if result.ConnectorType != tt.wantConnector {
				t.Errorf("ConnectorType = %q, want %q", result.ConnectorType, tt.wantConnector)
			}
			if result.PrepNote != tt.wantNote {
				t.Errorf("PrepNote = %q, want %q", result.PrepNote, tt.wantNote)
			}
			if tt.wantAdj != "" && result.PrepAdjective == "" {
				t.Errorf("PrepAdjective = %q, want %q", result.PrepAdjective, tt.wantAdj)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"beef broth", "Beef Broth"},
		{"all-purpose flour", "All-Purpose Flour"},
		{"BEEF BROTH", "Beef Broth"},
		{"chicken thighs", "Chicken Thighs"},
		{"", ""},
		{"single", "Single"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := titleCase(tt.input)
			if got != tt.want {
				t.Errorf("titleCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
