package detector

import (
	"strings"
)

// prepDescriptionDetector detects prep instructions or state adjectives in food names.
// Two sub-categories:
//   - "prep_instruction": user-performed actions (chopped, sliced, minced) → move to ingredient note
//   - "prep_attribute": product-altering state (diced, shredded, frozen) → keep in name
type prepDescriptionDetector struct{}

func newPrepDescriptionDetector() Detector {
	return &prepDescriptionDetector{}
}

func (d *prepDescriptionDetector) Name() string {
	return "prep_description"
}

func (d *prepDescriptionDetector) Check(f Food) []Issue {
	name := f.Name
	if name == "" {
		return nil
	}
	lower := strings.ToLower(name)

	var foundInstructions []string
	var foundAttributes []string

	for _, word := range userActionPrep {
		if strings.Contains(lower, word) {
			foundInstructions = append(foundInstructions, word)
		}
	}

	for _, word := range productAlteringPrep {
		if strings.Contains(lower, word) {
			foundAttributes = append(foundAttributes, word)
		}
	}

	var issues []Issue
	if len(foundInstructions) > 0 {
		issues = append(issues, Issue{
			Category: "prep_description",
			Severity: "medium",
			Message:  "Name contains user-performed prep instructions (should be ingredient note)",
			Details:  foundInstructions,
		})
	}
	// Attributes are fine in the name, so we don't flag them as issues.
	// They're detected for the normalizer's reference.
	if len(foundAttributes) > 0 {
		issues = append(issues, Issue{
			Category: "prep_attribute",
			Severity: "low",
			Message:  "Name contains product-altering prep attributes (keep in canonical name)",
			Details:  foundAttributes,
		})
	}

	return issues
}

// userActionPrep are prep words that represent actions the user performs.
// These should be moved to the ingredient note, not kept in the food name.
var userActionPrep = []string{
	"chopped", "sliced", "minced", "grated",
	"halved", "quartered", "diced",
	"peeled", "trimmed", "cubed",
	"crushed", // ambiguous: can be product state (crushed tomatoes)
}

// productAlteringPrep are prep words that describe the product itself.
// These should stay in the canonical food name.
var productAlteringPrep = []string{
	"frozen", "canned", "dried", "fresh",
	"boneless", "skinless", "with skin",
	"ground", "shredded",
	"condensed", "evaporated", "reduced",
	"smoked", "roasted", "toasted",
	"sweet", // e.g. "sweet pepper"
	"hot",   // e.g. "hot sauce"
}
