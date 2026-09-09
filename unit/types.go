// Package unit contains types for Tandoor Unit resources.
package unit

import "github.com/swedishborgie/go-tandoor/idref"

// Unit is a measurement unit (cup, tablespoon, etc.).
type Unit struct {
	ID           int    `json:"id,omitempty"`
	Name         string `json:"name"`
	PluralName   string `json:"plural_name,omitempty"`
	Description  string `json:"description,omitempty"`
	BaseUnit     string `json:"base_unit,omitempty"`
	OpenDataSlug string `json:"open_data_slug,omitempty"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full object (mirrors Tandoor's writable nested handling).
func (u *Unit) MarshalJSON() ([]byte, error) {
	type plain Unit
	return idref.MarshalJSON(u.ID, u.Name, plain(*u))
}

// UnmarshalJSON accepts either a bare integer or the full object.
func (u *Unit) UnmarshalJSON(data []byte) error {
	type plain Unit
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*u = Unit(p)
	return nil
}

// Ref is a minimal reference to a Unit (as returned by the API for nested
// references).
type Ref struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PluralName   string `json:"plural_name"`
	Description  string `json:"description"`
	BaseUnit     string `json:"base_unit"`
	OpenDataSlug string `json:"open_data_slug"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full object.
func (r *Ref) MarshalJSON() ([]byte, error) {
	type plain Ref
	return idref.MarshalJSON(r.ID, r.Name, plain(*r))
}

// UnmarshalJSON accepts either a bare integer or the full object.
func (r *Ref) UnmarshalJSON(data []byte) error {
	type plain Ref
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*r = Ref(p)
	return nil
}

// FoodRef is a minimal reference to a Food (as returned by the API for
// nested references).
type FoodRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full object.
func (f *FoodRef) MarshalJSON() ([]byte, error) {
	type plain FoodRef
	return idref.MarshalJSON(f.ID, f.Name, plain(*f))
}

// UnmarshalJSON accepts either a bare integer or the full object.
func (f *FoodRef) UnmarshalJSON(data []byte) error {
	type plain FoodRef
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*f = FoodRef(p)
	return nil
}

// Conversion defines a conversion between two units.
// The API accepts integer IDs in the request but returns full unit/food
// objects in the response.
type Conversion struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	BaseAmount      float64  `json:"base_amount"`
	BaseUnit        *Ref     `json:"base_unit"`
	ConvertedAmount float64  `json:"converted_amount"`
	ConvertedUnit   *Ref     `json:"converted_unit"`
	Food            *FoodRef `json:"food"`
	OpenDataSlug    string   `json:"open_data_slug"`
}
