// Package normalize cleans Tandoor food names into canonical form.
//
// Rules:
//   - Strip leading parenthetical quantities: "(10 1/2 oz. each) beef broth" → "Beef Broth"
//   - Strip leading numbers/fractions: "½ medium red onion" → "Red Onion"
//   - Strip leading connectors: "+ 1 tablespoon all-purpose flour" → "All-Purpose Flour"
//   - Keep product-altering prep: "diced tomatoes" → "Diced Tomatoes"
//   - Strip user-performed prep: "onion, sliced" → "Onion" (note: "sliced")
//   - Strip alternatives (keep first): "chicken broth or water" → "Chicken Broth"
//   - Strip can/ml references: "can black beans" → "Black Beans"
//   - Title Case output.
package normalize

import (
	"regexp"
	"strings"
)

// Result holds the output of normalizing a food name.
type Result struct {
	Canonical     string   // Title-cased canonical name.
	Stripped      []string // Pieces removed (quantities, connectors, etc.).
	PrepAdjective string   // Prep adjective to keep in name (e.g., "Diced").
	PrepNote      string   // Prep instruction to move to note (e.g., "sliced").
	Alternatives  []string // Alternative options (from "or" clauses).
	ConnectorType string   // "+", "-", "&", or "".
}

// Normalize takes a raw food name and returns a normalized Result.
func Normalize(name string) Result {
	var result Result
	var stripped []string

	raw := strings.TrimSpace(name)
	if raw == "" {
		return result
	}

	// 1. Detect and strip connector prefix.
	raw, conn := stripConnector(raw)
	if conn != "" {
		result.ConnectorType = conn
	}

	// 2. Strip leading parenthetical quantity/weight.
	raw, paren := stripParentheticalQuantity(raw)
	if paren != "" {
		stripped = append(stripped, paren)
	}

	// 3. Strip leading numbers, fractions, units.
	raw, leading := stripLeadingQuantity(raw)
	if leading != "" {
		stripped = append(stripped, leading)
	}

	// 4. Extract alternatives ("or" clauses).
	mainPart, alts := extractAlternatives(raw)
	raw = mainPart
	if len(alts) > 0 {
		result.Alternatives = alts
	}

	// 5. Strip packaging references (can, cans).
	raw, cans := stripPackaging(raw)
	if cans != "" {
		stripped = append(stripped, cans)
	}

	// 6. Detect prep words and classify them.
	prepNote, prepAdj := classifyPrep(raw)
	result.PrepNote = prepNote
	result.PrepAdjective = prepAdj

	// 7. Strip user-action prep from the name.
	if prepNote != "" {
		raw = stripUserPrep(raw)
	}

	// 8. Final cleanup: trim, collapse spaces, Title Case.
	result.Canonical = titleCase(cleanSpaces(raw))
	result.Stripped = stripped

	return result
}

// stripConnector removes a leading "+ ", "- ", or "& " prefix.
func stripConnector(s string) (string, string) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+") {
		return strings.TrimSpace(s[1:]), "+"
	}
	if strings.HasPrefix(s, "&") {
		return strings.TrimSpace(s[1:]), "&"
	}
	if strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "--") {
		return strings.TrimSpace(s[1:]), "-"
	}
	return s, ""
}

// stripParentheticalQuantity removes leading (...) like "(10 1/2 oz. each)".
func stripParentheticalQuantity(s string) (string, string) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "(") {
		return s, ""
	}
	closeIdx := strings.Index(s, ")")
	if closeIdx == -1 {
		return s, ""
	}
	paren := s[:closeIdx+1]
	if parentheticalQuantityRE.MatchString(paren) {
		return strings.TrimSpace(s[closeIdx+1:]), paren
	}
	return s, ""
}

// stripLeadingQuantity removes leading numeric/fraction/unit prefix.
func stripLeadingQuantity(s string) (string, string) {
	s = strings.TrimSpace(s)
	m := leadingQuantityRE.FindStringSubmatch(s)
	if m != nil {
		fullMatch := m[0]
		return strings.TrimSpace(s[len(fullMatch):]), fullMatch
	}
	return s, ""
}

// stripPackaging removes "can" or "cans" standalone words.
func stripPackaging(s string) (string, string) {
	matches := packagingRE.FindAllString(s, -1)
	if len(matches) == 0 {
		return s, ""
	}
	result := packagingRE.ReplaceAllString(s, "")
	return strings.TrimSpace(result), matches[0]
}

// extractAlternatives splits on " or " and returns the first option + alternatives.
func extractAlternatives(s string) (string, []string) {
	parts := alternativeRE.Split(s, -1)
	if len(parts) < 2 {
		return s, nil
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts[0], parts[1:]
}

// classifyPrep checks for prep words and returns (userActionNote, productAttribute).
func classifyPrep(s string) (string, string) {
	lower := strings.ToLower(s)
	var userActions []string
	var productAttrs []string

	for _, word := range userActionPrep {
		if strings.Contains(lower, word) {
			userActions = append(userActions, word)
		}
	}
	for _, word := range productAlteringPrep {
		if strings.Contains(lower, word) {
			productAttrs = append(productAttrs, word)
		}
	}

	var note string
	if len(userActions) > 0 {
		note = strings.Join(userActions, ", ")
	}
	var adj string
	if len(productAttrs) > 0 {
		adj = strings.Join(productAttrs, ", ")
	}
	return note, adj
}

// stripUserPrep removes user-action prep words from the name.
func stripUserPrep(s string) string {
	// Strip trailing comma-clause: "onion, sliced" → "onion"
	if idx := strings.LastIndex(s, ","); idx > 0 {
		head := strings.TrimSpace(s[:idx])
		tail := strings.TrimSpace(s[idx+1:])
		if isUserPrepWord(strings.ToLower(tail)) {
			return head
		}
	}
	// Strip standalone prep words.
	for _, word := range userActionPrep {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
		s = re.ReplaceAllString(s, "")
	}
	return cleanSpaces(s)
}

func isUserPrepWord(word string) bool {
	for _, w := range userActionPrep {
		if word == w {
			return true
		}
	}
	return false
}

// titleCase converts a string to Title Case.
func titleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Split on spaces, title each word.
	words := strings.Fields(s)
	for i := range words {
		words[i] = titleOne(words[i])
	}
	return strings.Join(words, " ")
}

// titleOne titles a single word, preserving hyphens.
func titleOne(word string) string {
	if word == "" {
		return ""
	}
	r := []rune(word)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	for i := 1; i < len(r); i++ {
		if r[i-1] == '-' {
			r[i] = []rune(strings.ToUpper(string(r[i])))[0]
			continue
		}
		r[i] = []rune(strings.ToLower(string(r[i])))[0]
	}
	return string(r)
}

// cleanSpaces collapses multiple spaces into one and trims.
func cleanSpaces(s string) string {
	return spaceRE.ReplaceAllString(s, " ")
}

// Word lists.

// userActionPrep are actions the user performs (move to ingredient note).
var userActionPrep = []string{
	"chopped", "sliced", "minced", "grated",
	"halved", "quartered",
	"peeled", "trimmed",
}

// productAlteringPrep describe the product itself (keep in name).
var productAlteringPrep = []string{
	"frozen", "dried", "fresh",
	"boneless", "skinless",
	"ground", "shredded",
	"condensed", "evaporated",
	"smoked", "roasted", "toasted",
	"crushed",
	"diced",
}

// Regexes.

var parentheticalQuantityRE = regexp.MustCompile(`[½¾¼0-9/]`)

// leadingQuantityRE matches a leading number, fraction, or unicode fraction
// followed by optional filler words (medium, large, cup, etc.) and units.
var leadingQuantityRE = regexp.MustCompile(
	`(?i)^` +
		`[½¾¼\s0-9/]+` +
		`[\s,]*` +
		`(?:medium|large|small|whole|)?` +
		`\s*` +
		`(?:cups?|tbsp|tsp|tablespoons?|teaspoons?|lbs?|lb|oz\.?|g\b)?\s*`,
)

var packagingRE = regexp.MustCompile(`(?i)\bcan\s*(s)?\b`)

var alternativeRE = regexp.MustCompile(`(?i)\s+or\s+`)

var spaceRE = regexp.MustCompile(`\s+`)
