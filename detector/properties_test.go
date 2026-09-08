package detector

import "testing"

var expectedTypes = []PropertyTypeInfo{
	{ID: 5, Name: "Proteins"},
	{ID: 6, Name: "Fats"},
	{ID: 7, Name: "Carbohydrates"},
	{ID: 8, Name: "Calories"},
}

func TestMissingPropertiesDetector_AllPresent(t *testing.T) {
	d := NewMissingPropertiesDetector(expectedTypes)
	issues := d.Check(Food{
		Name:            "Beef Broth",
		PropertyTypeIDs: []int{8, 5, 6, 7},
	})
	if len(issues) != 0 {
		t.Errorf("all present: got %d issues, want 0", len(issues))
	}
}

func TestMissingPropertiesDetector_SomeMissing(t *testing.T) {
	d := NewMissingPropertiesDetector(expectedTypes)
	issues := d.Check(Food{
		Name:            "Beef Broth",
		PropertyTypeIDs: []int{5, 8},
	})
	if len(issues) != 1 {
		t.Fatalf("some missing: got %d issues, want 1", len(issues))
	}
	iss := issues[0]
	if iss.Category != "missing_properties" {
		t.Errorf("Category = %q, want missing_properties", iss.Category)
	}
	if iss.Severity != "medium" {
		t.Errorf("Severity = %q, want medium", iss.Severity)
	}
	wantMsg := "Missing 2 property types: Fats, Carbohydrates"
	if iss.Message != wantMsg {
		t.Errorf("Message = %q, want %q", iss.Message, wantMsg)
	}
	wantIDs := []int{6, 7}
	gotIDs, ok := iss.Details.([]int)
	if !ok {
		t.Fatalf("Details is %T, want []int", iss.Details)
	}
	if len(gotIDs) != len(wantIDs) {
		t.Fatalf("Details = %v, want %v", gotIDs, wantIDs)
	}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Errorf("Details[%d] = %d, want %d", i, gotIDs[i], wantIDs[i])
		}
	}
}

func TestMissingPropertiesDetector_EmptyExpected(t *testing.T) {
	d := NewMissingPropertiesDetector(nil)
	issues := d.Check(Food{Name: "Beef Broth", PropertyTypeIDs: []int{5, 6, 7, 8}})
	if len(issues) != 0 {
		t.Errorf("empty expected: got %d issues, want 0 (no-op)", len(issues))
	}
}

func TestMissingPropertiesDetector_NoAttachedProperties(t *testing.T) {
	d := NewMissingPropertiesDetector(expectedTypes)
	issues := d.Check(Food{Name: "Beef Broth"})
	if len(issues) != 1 {
		t.Fatalf("no attached: got %d issues, want 1", len(issues))
	}
	wantMsg := "Missing 4 property types: Proteins, Fats, Carbohydrates, Calories"
	if issues[0].Message != wantMsg {
		t.Errorf("Message = %q, want %q", issues[0].Message, wantMsg)
	}
	wantIDs := []int{5, 6, 7, 8}
	gotIDs, ok := issues[0].Details.([]int)
	if !ok {
		t.Fatalf("Details is %T, want []int", issues[0].Details)
	}
	if len(gotIDs) != len(wantIDs) {
		t.Fatalf("Details = %v, want %v", gotIDs, wantIDs)
	}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Errorf("Details[%d] = %d, want %d", i, gotIDs[i], wantIDs[i])
		}
	}
}

func TestMissingPropertiesDetector_FiresWithoutFDCID(t *testing.T) {
	// The detector fires regardless of FDC state: a food with no FDC ID
	// gets both no_fdc_id (from the built-in detector) and
	// missing_properties when run through a registry.
	d := NewMissingPropertiesDetector(expectedTypes)
	issues := d.Check(Food{Name: "Beef Broth", FDCID: nil})
	if len(issues) != 1 {
		t.Fatalf("no FDC ID: got %d issues, want 1", len(issues))
	}
	if issues[0].Category != "missing_properties" {
		t.Errorf("Category = %q, want missing_properties", issues[0].Category)
	}
}

func TestMissingPropertiesDetector_Name(t *testing.T) {
	d := NewMissingPropertiesDetector(expectedTypes)
	if d.Name() != "missing_properties" {
		t.Errorf("Name = %q, want missing_properties", d.Name())
	}
}
