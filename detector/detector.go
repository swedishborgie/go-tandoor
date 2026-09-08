// Package detector provides pattern-based issue detection for Tandoor food names.
//
// Each Detector checks a food name (and optionally its FDC ID) and returns
// zero or more Issues. Detectors are registered in a Registry and can be
// run individually or in bulk.
//
//	registry := detector.NewRegistry()
//	issues := registry.CheckAll(name, fdcID)
package detector

// Issue represents a detected problem with a food entry.
type Issue struct {
	Category string // "quantity_prefix", "weight_unit", etc.
	Severity string // "high", "medium", "low"
	Message  string // Human-readable description.
	Details  any    // Arbitrary context (regex match, extracted quantity, etc.).
}

// Food is the minimal food shape needed for detection.
type Food struct {
	Name  string
	FDCID *int // nil means not set.

	// PropertyTypeIDs are the ids of property types currently attached to
	// the food (e.g. property types for calories, protein, ...). Populated
	// from the food's properties by the caller; nil/empty means none.
	PropertyTypeIDs []int
}

// Detector checks a food and returns zero or more issues.
type Detector interface {
	Name() string
	Check(f Food) []Issue
}

// Registry holds all detectors and provides bulk checking.
type Registry struct {
	detectors []Detector
}

// NewRegistry creates a registry with all built-in detectors registered.
func NewRegistry() *Registry {
	return &Registry{
		detectors: []Detector{
			newQuantityPrefixDetector(),
			newWeightUnitDetector(),
			newCanInNameDetector(),
			newMLInNameDetector(),
			newPrepDescriptionDetector(),
			newConnectorStartDetector(),
			newAlternativeInNameDetector(),
			newNoFDCIDDetector(),
		},
	}
}

// Add registers an additional detector.
func (r *Registry) Add(d Detector) {
	r.detectors = append(r.detectors, d)
}

// CheckAll runs all detectors against a food and returns all issues found.
func (r *Registry) CheckAll(f Food) []Issue {
	var issues []Issue
	for _, d := range r.detectors {
		issues = append(issues, d.Check(f)...)
	}
	return issues
}

// ByCategory returns the detector for the given category, or nil.
func (r *Registry) ByCategory(cat string) Detector {
	for _, d := range r.detectors {
		if d.Name() == cat {
			return d
		}
	}
	return nil
}

// Categories returns all registered category names.
func (r *Registry) Categories() []string {
	cats := make([]string, len(r.detectors))
	for i, d := range r.detectors {
		cats[i] = d.Name()
	}
	return cats
}

// filterByName runs only detectors matching the given category names.
// If cats is empty or nil, all detectors are run.
func (r *Registry) filterByName(cats []string) []Detector {
	if len(cats) == 0 {
		return r.detectors
	}
	set := make(map[string]struct{}, len(cats))
	for _, c := range cats {
		set[c] = struct{}{}
	}
	var filtered []Detector
	for _, d := range r.detectors {
		if _, ok := set[d.Name()]; ok {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// CheckCategories runs only the specified detectors.
func (r *Registry) CheckCategories(f Food, cats []string) []Issue {
	var issues []Issue
	for _, d := range r.filterByName(cats) {
		issues = append(issues, d.Check(f)...)
	}
	return issues
}
