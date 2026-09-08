package detector

import (
	"fmt"
	"strings"
)

// PropertyTypeInfo is the minimal property-type shape needed for the
// missing_properties detector. The expected set is enumerated from the
// server (GET /api/property-type/) so the detector package stays pure —
// it never imports the property package or a client.
type PropertyTypeInfo struct {
	ID   int
	Name string
}

// missingPropertiesDetector flags foods that are missing expected property
// types (e.g. a food with only Calories but no Proteins). It fires
// regardless of FDC state: a food without an FDC ID gets both no_fdc_id and
// missing_properties, since the fix path is the same.
type missingPropertiesDetector struct {
	expected []PropertyTypeInfo
}

// NewMissingPropertiesDetector returns a detector that reports expected
// property types absent from f.PropertyTypeIDs. It emits one issue per food
// listing all missing types. An empty expected slice makes the detector a
// no-op (safe degradation when the server enumeration fails).
func NewMissingPropertiesDetector(expected []PropertyTypeInfo) Detector {
	return &missingPropertiesDetector{expected: expected}
}

func (d *missingPropertiesDetector) Name() string {
	return "missing_properties"
}

func (d *missingPropertiesDetector) Check(f Food) []Issue {
	if len(d.expected) == 0 {
		return nil
	}
	attached := make(map[int]struct{}, len(f.PropertyTypeIDs))
	for _, id := range f.PropertyTypeIDs {
		attached[id] = struct{}{}
	}

	var missingIDs []int
	var missingNames []string
	for _, t := range d.expected {
		if _, ok := attached[t.ID]; !ok {
			missingIDs = append(missingIDs, t.ID)
			missingNames = append(missingNames, t.Name)
		}
	}
	if len(missingIDs) == 0 {
		return nil
	}

	return []Issue{{
		Category: "missing_properties",
		Severity: "medium",
		Message:  fmt.Sprintf("Missing %d property types: %s", len(missingIDs), strings.Join(missingNames, ", ")),
		Details:  missingIDs,
	}}
}
