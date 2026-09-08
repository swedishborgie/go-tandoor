package detector

import "regexp"

// canInNameDetector detects can/cans references in food names.
//   - "can diced tomatoes" → "diced tomatoes"
//   - These are packaging info, not part of the food identity.
type canInNameDetector struct{}

func newCanInNameDetector() Detector {
	return &canInNameDetector{}
}

func (d *canInNameDetector) Name() string {
	return "can_in_name"
}

func (d *canInNameDetector) Check(f Food) []Issue {
	name := f.Name
	if name == "" {
		return nil
	}

	if canRE.MatchString(name) {
		return []Issue{{
			Category: "can_in_name",
			Severity: "medium",
			Message:  "Name contains 'can' or 'cans' (packaging reference)",
			Details:  canRE.FindString(name),
		}}
	}

	return nil
}

var canRE = regexp.MustCompile(`(?i)\bcan\b`)
