package detector

import "regexp"

// quantityPrefixDetector detects quantity prefixes like:
//   - "(10 1/2 oz. each) beef broth"
//   - "(398mL) can diced tomatoes"
//   - "(10 g) kosher salt"
//   - "(10¾ oz.) condensed cream..."
//   - Leading numbers: "1/2 cup chopped onion"
//   - Leading fractions: "½ medium red onion"
type quantityPrefixDetector struct{}

func newQuantityPrefixDetector() Detector {
	return &quantityPrefixDetector{}
}

func (d *quantityPrefixDetector) Name() string {
	return "quantity_prefix"
}

func (d *quantityPrefixDetector) Check(f Food) []Issue {
	name := f.Name
	if name == "" {
		return nil
	}

	// Pattern 1: leading parenthetical quantity like "(10 1/2 oz. each) beef broth"
	// Matches: (...) at start where content contains digits, fractions, or units.
	if parentheticalRE.MatchString(name) {
		return []Issue{{
			Category: "quantity_prefix",
			Severity: "high",
			Message:  "Name starts with parenthetical quantity/weight",
			Details:  extractMatch(parentheticalRE, name),
		}}
	}

	// Pattern 2: leading number or fraction like "1/2 cup..." or "2 lbs..."
	if leadingNumberRE.MatchString(name) {
		return []Issue{{
			Category: "quantity_prefix",
			Severity: "high",
			Message:  "Name starts with a numeric quantity",
			Details:  extractMatch(leadingNumberRE, name),
		}}
	}

	// Pattern 3: leading unicode fraction like "½ medium red onion"
	if leadingFractionRE.MatchString(name) {
		return []Issue{{
			Category: "quantity_prefix",
			Severity: "high",
			Message:  "Name starts with a unicode fraction",
			Details:  extractMatch(leadingFractionRE, name),
		}}
	}

	return nil
}

var (
	// parentheticalRE matches leading (...) with content that looks like a quantity.
	parentheticalRE = regexp.MustCompile(`^\(\s*[½¾¼0-9/]`)
	// leadingNumberRE matches names starting with a digit or fraction.
	leadingNumberRE = regexp.MustCompile(`^\s*[0-9]+`)
	// leadingFractionRE matches names starting with unicode fractions.
	leadingFractionRE = regexp.MustCompile(`^\s*[½¾¼]`)
)

func extractMatch(re *regexp.Regexp, s string) string {
	if m := re.FindString(s); m != "" {
		return m
	}
	return ""
}
