// Package food contains types for Tandoor Food resources.
package food

import "github.com/swedishborgie/go-tandoor/property"

// Food is a food item that can be used in recipes and shopping lists.
type Food struct {
	ID                   int                 `json:"id,omitempty"`
	Name                 string              `json:"name"`
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
