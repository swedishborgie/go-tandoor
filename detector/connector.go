package detector

import "strings"

// connectorStartDetector detects connector prefixes in food names.
//   - "+ " (compound amount): "1 cup + 1 tablespoon flour" → second line starts with "+"
//   - "- " (range amount): "1 lb - 1 1/4 lb chicken thighs"
//   - "& " (multiple foods): "1 red & 1 orange sweet pepper"
type connectorStartDetector struct{}

func newConnectorStartDetector() Detector {
	return &connectorStartDetector{}
}

func (d *connectorStartDetector) Name() string {
	return "connector_start"
}

func (d *connectorStartDetector) Check(f Food) []Issue {
	name := strings.TrimSpace(f.Name)
	if name == "" {
		return nil
	}

	var connector string
	var fixable string

	switch {
	case strings.HasPrefix(name, "+"):
		connector = "+"
		fixable = "rename+merge"
	case strings.HasPrefix(name, "-") && !strings.HasPrefix(name, "--"):
		connector = "-"
		fixable = "rename+merge"
	case strings.HasPrefix(name, "&"):
		connector = "&"
		fixable = "requires_split"
	}

	if connector != "" {
		return []Issue{{
			Category: "connector_start",
			Severity: "high",
			Message:  "Name starts with connector prefix " + connector,
			Details: map[string]string{
				"connector": connector,
				"action":    fixable,
			},
		}}
	}

	return nil
}

// alternativeInNameDetector detects "or" alternatives baked into the food name.
//   - "chicken broth or water" → name: "chicken broth", note: "or water"
type alternativeInNameDetector struct{}

func newAlternativeInNameDetector() Detector {
	return &alternativeInNameDetector{}
}

func (d *alternativeInNameDetector) Name() string {
	return "alternative_in_name"
}

func (d *alternativeInNameDetector) Check(f Food) []Issue {
	name := f.Name
	if name == "" {
		return nil
	}

	// Look for " or " with word boundary-like context.
	// Avoid matching "or" inside words like "flavor".
	idx := strings.Index(strings.ToLower(name), " or ")
	if idx == -1 {
		return nil
	}

	return []Issue{{
		Category: "alternative_in_name",
		Severity: "medium",
		Message:  "Name contains 'or' alternative",
		Details:  name[idx:],
	}}
}
