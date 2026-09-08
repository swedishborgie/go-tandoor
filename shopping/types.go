// Package shopping contains types for Tandoor Shopping List resources.
package shopping

import (
	"time"

	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/swedishborgie/go-tandoor/unit"
)

// List represents a user's shopping list.
type List struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// ListEntry is an individual item in a shopping list.
type ListEntry struct {
	ID int `json:"id,omitempty"`

	// Food is the food reference (can be nil for headers).
	Food *food.Shopping `json:"food,omitempty"`

	// Unit is the measurement unit.
	Unit *unit.Unit `json:"unit,omitempty"`

	// Lists are the lists this entry belongs to.
	Lists []List `json:"shopping_lists,omitempty"`

	// ListRecipeData is the shopping list recipe reference (read-only).
	ListRecipeData *ListRecipe `json:"list_recipe_data,omitempty"`

	// Amount is the quantity.
	Amount float64 `json:"amount,omitempty"`

	// Note is a free-text note.
	Note string `json:"note,omitempty"`

	// Order controls display order.
	Order int `json:"order,omitempty"`

	// IsHeader marks this as a section header.
	IsHeader bool `json:"is_header,omitempty"`

	// NoAmount indicates no quantity should be shown.
	NoAmount bool `json:"no_amount,omitempty"`

	// OriginalText is the raw ingredient text.
	OriginalText string `json:"original_text,omitempty"`

	// Checked marks the item as purchased.
	Checked bool `json:"checked,omitempty"`

	// CompletedAt is when the item was checked off.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// CreatedBy is the user who created this entry.
	CreatedBy recipe.User `json:"created_by,omitempty"`

	// UpdatedAt is the last update time.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ListRecipe links a recipe to a shopping list entry.
type ListRecipe struct {
	ID int `json:"id,omitempty"`

	// Name is a convenience field for the recipe name.
	Name string `json:"name,omitempty"`

	// Recipe is the recipe reference.
	Recipe *recipe.Overview `json:"recipe_data,omitempty"`

	// MealPlan is the associated meal plan (read-only).
	MealPlan *any `json:"meal_plan_data,omitempty"`

	// Servings overrides the default servings.
	Servings float64 `json:"servings,omitempty"`

	// CreatedBy is the user who added this.
	CreatedBy recipe.User `json:"created_by,omitempty"`
}

// ListEntryBulk is the request body for bulk update operations.
type ListEntryBulk struct {
	IDs []int `json:"ids"`

	// Checked sets the checked state (nil to leave unchanged).
	Checked *bool `json:"checked,omitempty"`

	// Timestamp is returned by the API (read-only).
	Timestamp time.Time `json:"timestamp"`

	// ListsAdd adds entries to these lists.
	ListsAdd []int `json:"shopping_lists_add,omitempty"`

	// ListsRemove removes entries from these lists.
	ListsRemove []int `json:"shopping_lists_remove,omitempty"`

	// ListsSet replaces all list associations.
	ListsSet []int `json:"shopping_lists_set,omitempty"`

	// ListsRemoveAll removes entries from all lists.
	ListsRemoveAll bool `json:"shopping_lists_remove_all"`
}

// ListEntryBulkCreate is the request body for bulk creating entries.
type ListEntryBulkCreate struct {
	Entries []ListEntryCreate `json:"entries"`
	ListIDs []int             `json:"shopping_lists_ids,omitempty"`
}

// ListEntryCreate is the simplified entry for bulk create.
type ListEntryCreate struct {
	FoodID *int    `json:"food"`
	UnitID *int    `json:"unit"`
	Amount float64 `json:"amount"`
	Note   string  `json:"note"`
}
