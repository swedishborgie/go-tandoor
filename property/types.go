// Package property contains types for Tandoor Property resources.
package property

// Type defines a category of food property (e.g. "calories", "protein", "sodium").
type Type struct {
	ID           int    `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Unit         string `json:"unit,omitempty"`
	Description  string `json:"description,omitempty"`
	Order        int    `json:"order,omitempty"`
	OpenDataSlug string `json:"open_data_slug,omitempty"`
	FDCID        *int   `json:"fdc_id,omitempty"`
}

// FoodRef is a minimal food reference used in property payloads.
type FoodRef struct {
	ID int `json:"id"`
}

// Property links a food to a property type with a measured amount.
type Property struct {
	ID int `json:"id"`

	// Food is the food this property belongs to.
	// Tandoor expects a nested object {"id": X}, not a bare integer.
	Food *FoodRef `json:"food,omitempty"`

	// Type is the property category (calories, protein, etc.).
	Type Type `json:"property_type"`

	// PropertyAmount is the measured value (e.g. 250 kcal, 12g protein).
	// Can be null for boolean/categorical properties.
	PropertyAmount *float64 `json:"property_amount"`
}
