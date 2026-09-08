package detector

import "regexp"

// mlInNameDetector detects mL/ml references in food names.
//   - "(127mL) can chopped green chilis" → "Chopped Green Chilis"
//   - Volume references in parentheticals are quantity info, not food identity.
type mlInNameDetector struct{}

func newMLInNameDetector() Detector {
	return &mlInNameDetector{}
}

func (d *mlInNameDetector) Name() string {
	return "ml_in_name"
}

func (d *mlInNameDetector) Check(f Food) []Issue {
	name := f.Name
	if name == "" {
		return nil
	}

	if mlRE.MatchString(name) {
		return []Issue{{
			Category: "ml_in_name",
			Severity: "medium",
			Message:  "Name contains mL/ml volume reference",
			Details:  mlRE.FindString(name),
		}}
	}

	return nil
}

var mlRE = regexp.MustCompile(`(?i)(?:[0-9]m[Ll]|\bml\b)`)
