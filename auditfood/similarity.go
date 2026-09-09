// Package auditfood contains the shared food-audit logic used by both the
// tandoor-cli commands and the MCP tools: food ensure, property
// attach, name auditing (inspect/fix), duplicate detection, FDC property
// import, and auto unit conversions.
//
// Conventions (Title Case, prep-word stripping, Jaccard thresholds, the
// per-100 g property basis) are data-driven where the server provides them
// (expected property types are enumerated from the instance) and exposed as
// options here rather than hardcoded.
package auditfood

import "strings"

// WordSet splits a name into a set of lowercased words.
func WordSet(name string) map[string]struct{} {
	words := strings.Fields(strings.ToLower(name))
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		set[w] = struct{}{}
	}
	return set
}

// Jaccard computes the Jaccard similarity between two word sets (0..1).
// Two empty sets have similarity 0.
func Jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	intersection := 0
	for w := range a {
		if _, ok := b[w]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// JaccardSimilarity is word-level Jaccard similarity between two names.
// Two empty names are considered identical (1.0).
func JaccardSimilarity(a, b string) float64 {
	setA := WordSet(strings.ToLower(a))
	setB := WordSet(strings.ToLower(b))
	if len(setA) == 0 && len(setB) == 0 {
		return 1.0
	}
	return Jaccard(setA, setB)
}
