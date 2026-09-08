package detector

import "regexp"

// weightUnitDetector detects weight/unit references baked into food names.
//   - Contains "oz", "lb", "ounces", "grams", "g)"
//   - These should be in the ingredient amount, not the food name.
type weightUnitDetector struct{}

func newWeightUnitDetector() Detector {
	return &weightUnitDetector{}
}

func (d *weightUnitDetector) Name() string {
	return "weight_unit"
}

func (d *weightUnitDetector) Check(f Food) []Issue {
	name := f.Name
	if name == "" {
		return nil
	}

	// Skip names that are purely a connector prefix (handled by connector detector).
	// We still want to flag weight units in non-connector names.

	var matches []string
	for _, re := range weightUnitREs {
		if m := re.FindString(name); m != "" {
			matches = append(matches, m)
		}
	}

	if len(matches) > 0 {
		return []Issue{{
			Category: "weight_unit",
			Severity: "high",
			Message:  "Name contains weight or unit references",
			Details:  matches,
		}}
	}

	return nil
}

var weightUnitREs = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:^|\b)oz\.?(?:\b|$)`),   // oz, oz.
	regexp.MustCompile(`(?i)(?:^|\b)lb(?:\b|$)`),      // lb
	regexp.MustCompile(`(?i)(?:^|\b)lbs?(?:\b|$)`),    // lbs
	regexp.MustCompile(`(?i)(?:^|\b)ounces?(?:\b|$)`), // ounce, ounces
	regexp.MustCompile(`(?i)(?:^|\b)grams?(?:\b|$)`),  // gram, grams
	regexp.MustCompile(`(?i)[0-9]\s*g\s*\)`),          // g) - parenthetical gram ref (e.g., "10 g)")
	regexp.MustCompile(`(?i)(?:^|\b)kg(?:\b|$)`),      // kg
	regexp.MustCompile(`(?i)[0-9]m[Ll]`),              // mL, ml after digits (e.g., 127mL)
	regexp.MustCompile(`(?i)(?:^|\b)ml(?:\b|$)`),      // ml standalone
}
