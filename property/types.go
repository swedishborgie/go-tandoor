// Package property contains types for Tandoor Property resources.
package property

import "github.com/swedishborgie/go-tandoor/idref"

// Type defines a category of food property (e.g. "calories", "protein",
// "sodium").
type Type struct {
	ID           int    `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Unit         string `json:"unit,omitempty"`
	Description  string `json:"description,omitempty"`
	Order        int    `json:"order,omitempty"`
	OpenDataSlug string `json:"open_data_slug,omitempty"`
	FDCID        *int   `json:"fdc_id,omitempty"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full object (mirrors Tandoor's writable nested handling).
func (t *Type) MarshalJSON() ([]byte, error) {
	type plain Type
	return idref.MarshalJSON(t.ID, t.Name, plain(*t))
}

// UnmarshalJSON accepts either a bare integer or the full object.
func (t *Type) UnmarshalJSON(data []byte) error {
	type plain Type
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*t = Type(p)
	return nil
}

// FoodRef is a minimal food reference used in property payloads.
type FoodRef struct {
	ID int `json:"id"`
}

// MarshalJSON marshals as a bare integer when an ID is set.
func (f *FoodRef) MarshalJSON() ([]byte, error) {
	type plain FoodRef
	return idref.MarshalJSON(f.ID, "", plain(*f))
}

// UnmarshalJSON accepts either a bare integer or an object with an id.
func (f *FoodRef) UnmarshalJSON(data []byte) error {
	type plain FoodRef
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*f = FoodRef(p)
	return nil
}

// Property links a food to a property type with a measured amount.
type Property struct {
	ID int `json:"id,omitempty"`

	// Food is the food this property belongs to.
	Food *FoodRef `json:"food,omitempty"`

	// Type is the property category (calories, protein, etc.).
	Type Type `json:"property_type"`

	// PropertyAmount is the measured value (e.g. 250 kcal, 12g protein).
	// Can be null for boolean/categorical properties.
	PropertyAmount *float64 `json:"property_amount"`
}
