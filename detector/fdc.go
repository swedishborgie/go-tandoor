package detector

// noFDCIDDetector flags foods that have no FDC ID set.
type noFDCIDDetector struct{}

func newNoFDCIDDetector() Detector {
	return &noFDCIDDetector{}
}

func (d *noFDCIDDetector) Name() string {
	return "no_fdc_id"
}

func (d *noFDCIDDetector) Check(f Food) []Issue {
	if f.FDCID == nil {
		return []Issue{{
			Category: "no_fdc_id",
			Severity: "low",
			Message:  "Food has no USDA FDC ID set",
			Details:  nil,
		}}
	}
	return nil
}
