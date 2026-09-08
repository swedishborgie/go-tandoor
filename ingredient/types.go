// Package ingredient contains types for Tandoor Ingredient resources.
package ingredient

import "encoding/json"

// Ingredient is an ingredient as returned by the standalone /api/ingredient/
// endpoint.  Unlike the nested Ingredient in recipe/types.go (which has
// integer food/unit IDs), this endpoint returns full food and unit objects.
type Ingredient struct {
	ID            int               `json:"id,omitempty"`
	FoodID        interface{}       `json:"food,omitempty"`
	UnitID        interface{}       `json:"unit,omitempty"`
	Food          *Food             `json:"-"`
	Unit          *Unit             `json:"-"`
	Amount        float64           `json:"amount,omitempty"`
	Note          string            `json:"note,omitempty"`
	Order         int               `json:"order,omitempty"`
	IsHeader      bool              `json:"is_header,omitempty"`
	NoAmount      bool              `json:"no_amount,omitempty"`
	OriginalText  string            `json:"original_text,omitempty"`
	Checked       bool              `json:"checked,omitempty"`
	UsedInRecipes []RecipeReference `json:"used_in_recipes,omitempty"`
}

// UnmarshalJSON custom unmarshals Ingredient to handle food/unit as int or object.
func (i *Ingredient) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID            int               `json:"id,omitempty"`
		Food          json.RawMessage   `json:"food,omitempty"`
		Unit          json.RawMessage   `json:"unit,omitempty"`
		Amount        float64           `json:"amount,omitempty"`
		Note          string            `json:"note,omitempty"`
		Order         int               `json:"order,omitempty"`
		IsHeader      bool              `json:"is_header,omitempty"`
		NoAmount      bool              `json:"no_amount,omitempty"`
		OriginalText  string            `json:"original_text,omitempty"`
		Checked       bool              `json:"checked,omitempty"`
		UsedInRecipes []RecipeReference `json:"used_in_recipes,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	i.ID = aux.ID
	i.Amount = aux.Amount
	i.Note = aux.Note
	i.Order = aux.Order
	i.IsHeader = aux.IsHeader
	i.NoAmount = aux.NoAmount
	i.OriginalText = aux.OriginalText
	i.Checked = aux.Checked
	i.UsedInRecipes = aux.UsedInRecipes
	// food can be int or object
	if len(aux.Food) > 0 {
		var fid int
		if json.Unmarshal(aux.Food, &fid) == nil {
			i.FoodID = fid
		} else {
			var f Food
			if json.Unmarshal(aux.Food, &f) == nil {
				i.Food = &f
				i.FoodID = f.ID
			}
		}
	}
	// unit can be int or object
	if len(aux.Unit) > 0 {
		var uid int
		if json.Unmarshal(aux.Unit, &uid) == nil {
			i.UnitID = uid
		} else {
			var u Unit
			if json.Unmarshal(aux.Unit, &u) == nil {
				i.Unit = &u
				i.UnitID = u.ID
			}
		}
	}
	return nil
}

// MarshalJSON custom marshals Ingredient to send food/unit as IDs.
func (i Ingredient) MarshalJSON() ([]byte, error) {
	type Alias Ingredient
	a := struct {
		Alias
		Food json.RawMessage `json:"food,omitempty"`
		Unit json.RawMessage `json:"unit,omitempty"`
	}{
		Alias: Alias(i),
	}
	// marshal food as ID
	if i.Food != nil {
		foodID, err := json.Marshal(i.Food.ID)
		if err == nil {
			a.Food = foodID
		}
	} else if id, ok := i.FoodID.(int); ok {
		foodID, err := json.Marshal(id)
		if err == nil {
			a.Food = foodID
		}
	}
	// marshal unit as ID
	if i.Unit != nil {
		unitID, err := json.Marshal(i.Unit.ID)
		if err == nil {
			a.Unit = unitID
		}
	} else if id, ok := i.UnitID.(int); ok {
		unitID, err := json.Marshal(id)
		if err == nil {
			a.Unit = unitID
		}
	}
	return json.Marshal(a)
}

// Food is the food reference within an ingredient.
type Food struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Unit is the unit reference within an ingredient.
type Unit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RecipeReference is a minimal recipe reference (used in UsedInRecipes).
type RecipeReference struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// FoodIDValue returns the food ID, or nil if the food is not set.
func (i *Ingredient) FoodIDValue() *int {
	if i.Food == nil {
		return nil
	}
	return &i.Food.ID
}

// UnitIDValue returns the unit ID, or nil if the unit is not set.
func (i *Ingredient) UnitIDValue() *int {
	if i.Unit == nil {
		return nil
	}
	return &i.Unit.ID
}

// Simple is a minimal ingredient reference for use in
// nested structures.
type Simple struct {
	ID           int     `json:"id"`
	FoodID       *int    `json:"food"`
	UnitID       *int    `json:"unit"`
	Amount       float64 `json:"amount"`
	Note         string  `json:"note"`
	Order        int     `json:"order"`
	IsHeader     bool    `json:"is_header"`
	NoAmount     bool    `json:"no_amount"`
	OriginalText string  `json:"original_text"`
	Checked      bool    `json:"checked"`
}
