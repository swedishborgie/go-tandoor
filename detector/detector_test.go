package detector

import "testing"

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	cats := r.Categories()
	expected := []string{
		"quantity_prefix",
		"weight_unit",
		"can_in_name",
		"ml_in_name",
		"prep_description",
		"connector_start",
		"alternative_in_name",
		"no_fdc_id",
	}
	if len(cats) != len(expected) {
		t.Fatalf("got %d categories, want %d", len(cats), len(expected))
	}
	for i, want := range expected {
		if cats[i] != want {
			t.Errorf("cats[%d] = %q, want %q", i, cats[i], want)
		}
	}
}

func TestDetector_QuantityPrefix(t *testing.T) {
	d := NewRegistry().ByCategory("quantity_prefix")
	tests := []struct {
		name      string
		input     string
		wantIssue bool
	}{
		{"parenthetical oz", "(10 1/2 oz. each) beef broth", true},
		{"parenthetical g", "(10 g) kosher salt", true},
		{"parenthetical mL", "(398mL) can diced tomatoes", true},
		{"parenthetical fraction", "(10¾ oz.) cream soup", true},
		{"leading number", "1/2 cup flour", true},
		{"leading fraction", "½ medium red onion", true},
		{"clean name", "beef broth", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := d.Check(Food{Name: tt.input})
			hasIssue := len(issues) > 0
			if hasIssue != tt.wantIssue {
				t.Errorf("got %d issues (hasIssue=%v), want %v", len(issues), hasIssue, tt.wantIssue)
			}
		})
	}
}

func TestDetector_WeightUnit(t *testing.T) {
	d := NewRegistry().ByCategory("weight_unit")
	tests := []struct {
		name      string
		input     string
		wantIssue bool
	}{
		{"oz in name", "(10 1/2 oz. each) beef broth", true},
		{"lbs in name", "1 ½ lbs boneless chicken breast", true},
		{"clean name", "beef broth", false},
		{"grams", "(10 g) kosher salt", true},
		{"mL", "(127mL) can chopped chilis", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := d.Check(Food{Name: tt.input})
			hasIssue := len(issues) > 0
			if hasIssue != tt.wantIssue {
				t.Errorf("got %d issues (hasIssue=%v), want %v", len(issues), hasIssue, tt.wantIssue)
			}
		})
	}
}

func TestDetector_CanInName(t *testing.T) {
	d := NewRegistry().ByCategory("can_in_name")
	tests := []struct {
		name      string
		input     string
		wantIssue bool
	}{
		{"can in name", "(398mL) can diced tomatoes", true},
		{"no can", "diced tomatoes", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := d.Check(Food{Name: tt.input})
			hasIssue := len(issues) > 0
			if hasIssue != tt.wantIssue {
				t.Errorf("got %d issues, want hasIssue=%v", len(issues), tt.wantIssue)
			}
		})
	}
}

func TestDetector_ConnectorStart(t *testing.T) {
	d := NewRegistry().ByCategory("connector_start")
	tests := []struct {
		name      string
		input     string
		wantIssue bool
		connector string
	}{
		{"plus connector", "+ 1 tablespoon flour", true, "+"},
		{"minus connector", "- 1 1/4 lb chicken", true, "-"},
		{"ampersand connector", "& 1 orange pepper", true, "&"},
		{"clean name", "beef broth", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := d.Check(Food{Name: tt.input})
			hasIssue := len(issues) > 0
			if hasIssue != tt.wantIssue {
				t.Errorf("hasIssue = %v, want %v", hasIssue, tt.wantIssue)
			}
			if hasIssue && tt.wantIssue {
				details, ok := issues[0].Details.(map[string]string)
				if !ok {
					t.Errorf("Details not map[string]string")
				} else if details["connector"] != tt.connector {
					t.Errorf("connector = %q, want %q", details["connector"], tt.connector)
				}
			}
		})
	}
}

func TestDetector_AlternativeInName(t *testing.T) {
	d := NewRegistry().ByCategory("alternative_in_name")
	tests := []struct {
		name      string
		input     string
		wantIssue bool
	}{
		{"with alternative", "chicken broth or water", true},
		{"no alternative", "chicken broth", false},
		{"number with alt", "1 cup chicken broth or water", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := d.Check(Food{Name: tt.input})
			hasIssue := len(issues) > 0
			if hasIssue != tt.wantIssue {
				t.Errorf("hasIssue = %v, want %v", hasIssue, tt.wantIssue)
			}
		})
	}
}

func TestDetector_NoFDCID(t *testing.T) {
	d := NewRegistry().ByCategory("no_fdc_id")
	// No FDC ID set.
	issues := d.Check(Food{Name: "beef broth"})
	if len(issues) != 1 {
		t.Errorf("no FDC ID: got %d issues, want 1", len(issues))
	}
	// FDC ID is set.
	fdcID := 189196
	issues = d.Check(Food{Name: "beef broth", FDCID: &fdcID})
	if len(issues) != 0 {
		t.Errorf("FDC ID set: got %d issues, want 0", len(issues))
	}
}

func TestRegistry_CheckAll(t *testing.T) {
	r := NewRegistry()
	// This food should trigger multiple issues: quantity_prefix, weight_unit, can_in_name, ml_in_name, no_fdc_id
	issues := r.CheckAll(Food{Name: "(398mL) can diced tomatoes"})
	categories := make(map[string]bool)
	for _, issue := range issues {
		categories[issue.Category] = true
	}
	// Should have at least these categories.
	expected := []string{"quantity_prefix", "weight_unit", "can_in_name", "ml_in_name", "no_fdc_id"}
	for _, cat := range expected {
		if !categories[cat] {
			t.Errorf("missing expected category %q (got %v)", cat, categories)
		}
	}
}

func TestRegistry_ByCategory(t *testing.T) {
	r := NewRegistry()
	d := r.ByCategory("quantity_prefix")
	if d == nil {
		t.Fatal("ByCategory(quantity_prefix) returned nil")
	}
	if d.Name() != "quantity_prefix" {
		t.Errorf("Name = %q, want quantity_prefix", d.Name())
	}
	// Non-existent category.
	d = r.ByCategory("nonexistent")
	if d != nil {
		t.Errorf("ByCategory(nonexistent) returned non-nil")
	}
}

func TestRegistry_CheckCategories(t *testing.T) {
	r := NewRegistry()
	issues := r.CheckCategories(Food{Name: "(398mL) can diced tomatoes"}, []string{"quantity_prefix", "can_in_name"})
	for _, issue := range issues {
		switch issue.Category {
		case "quantity_prefix", "can_in_name":
			// expected
		default:
			t.Errorf("unexpected category %q", issue.Category)
		}
	}
}
