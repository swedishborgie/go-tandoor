// Package unit contains types for Tandoor Unit resources.
package unit

// Unit is a measurement unit (cup, tablespoon, etc.).
type Unit struct {
	ID           int    `json:"id,omitempty"`
	Name         string `json:"name"`
	PluralName   string `json:"plural_name,omitempty"`
	Description  string `json:"description,omitempty"`
	BaseUnit     string `json:"base_unit,omitempty"`
	OpenDataSlug string `json:"open_data_slug,omitempty"`
}

// Ref is a minimal reference to a Unit (as returned by the API for nested references).
type Ref struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PluralName   string `json:"plural_name"`
	Description  string `json:"description"`
	BaseUnit     string `json:"base_unit"`
	OpenDataSlug string `json:"open_data_slug"`
}

// FoodRef is a minimal reference to a Food (as returned by the API for nested references).
type FoodRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Conversion defines a conversion between two units.
// The API accepts integer IDs in the request but returns full unit/food objects in the response.
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
