// Package food contains types for Tandoor Food resources.
package food

import (
	"github.com/swedishborgie/go-tandoor/idref"
	"github.com/swedishborgie/go-tandoor/property"
)

// Food is a food item that can be used in recipes and shopping lists.
type Food struct {
	ID int `json:"id,omitempty"`
	// Name omits when empty so partial updates (PATCH) do not send a blank
	// name; Tandoor rejects blank names on create/update.
	Name                 string              `json:"name,omitempty"`
	PluralName           string              `json:"plural_name,omitempty"`
	Description          string              `json:"description,omitempty"`
	Shopping             string              `json:"shopping,omitempty"`
	Properties           []property.Property `json:"properties,omitempty"`
	PropertiesFoodAmount *float64            `json:"properties_food_amount,omitempty"`
	PropertiesFoodUnit   any                 `json:"properties_food_unit,omitempty"` // unit id or object
	FDCID                *int                `json:"fdc_id,omitempty"`
	IgnoreShopping       bool                `json:"ignore_shopping,omitempty"`
	OpenDataSlug         string              `json:"open_data_slug,omitempty"`
	Parent               int                 `json:"parent,omitempty"`
	NumChild             int                 `json:"numchild,omitempty"`
	FullName             string              `json:"full_name,omitempty"`
	SubstituteOnHand     bool                `json:"substitute_onhand,omitempty"`
}

// Simple is a minimal food reference.
type Simple struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PluralName string `json:"plural_name"`
}

// BatchUpdate performs batch operations on multiple foods.
type BatchUpdate struct {
	Foods            []int `json:"foods"`
	Category         *int  `json:"category,omitempty"`
	SubstituteAdd    []int `json:"substitute_add,omitempty"`
	SubstituteRemove []int `json:"substitute_remove,omitempty"`
	SubstituteSet    []int `json:"substitute_set,omitempty"`
}

// Shopping updates shopping behavior for a food.
type Shopping struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	IgnoreShopping bool   `json:"ignore_shopping"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full object (mirrors Tandoor's writable nested handling).
func (s *Shopping) MarshalJSON() ([]byte, error) {
	type plain Shopping
	return idref.MarshalJSON(s.ID, s.Name, plain(*s))
}

// UnmarshalJSON accepts either a bare integer or the full object.
func (s *Shopping) UnmarshalJSON(data []byte) error {
	type plain Shopping
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*s = Shopping(p)
	return nil
}

// ShoppingUpdate requests a shopping update for a food.
type ShoppingUpdate struct {
	ShoppingListID int `json:"shopping_list"`
}

// InheritField describes an inherited field on a food.
type InheritField struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Field string `json:"field"`
}
