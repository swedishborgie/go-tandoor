// Package mealplan contains types for Tandoor Meal Planning resources.
package mealplan

import (
	"encoding/json"
	"time"

	"github.com/swedishborgie/go-tandoor/recipe"
)

// MealPlan represents a single entry in the meal planner.
type MealPlan struct {
	ID int `json:"id,omitempty"`

	// Title is an optional human-readable title.
	Title string `json:"title,omitempty"`

	// Recipe is the recipe reference (can be nil for free-text entries).
	Recipe *recipe.Overview `json:"recipe,omitempty"`

	// RecipeName is a read-only convenience field.
	RecipeName string `json:"recipe_name,omitempty"`

	// MealTypeID is the meal type ID for write operations.
	MealTypeID int `json:"meal_type,omitempty"`

	// MealType is the type of meal (breakfast, lunch, dinner, etc.).
	MealType *MealType `json:"-"`
	// MealTypeName is a read-only convenience field.
	MealTypeName string `json:"meal_type_name,omitempty"`

	// Note is a free-text note for the meal plan entry.
	Note string `json:"note,omitempty"`

	// NoteMarkdown is the note rendered as markdown.
	NoteMarkdown string `json:"note_markdown,omitempty"`

	// Servings overrides the default number of servings.
	Servings float64 `json:"servings,omitempty"`

	// Shopping indicates whether this meal plan is in a shopping list.
	Shopping bool `json:"shopping,omitempty"`

	// AddShopping adds the meal plan to the shopping list (write-only).
	AddShopping bool `json:"addshopping,omitempty"`

	// FromDate is the start date/time of the meal plan entry. A pointer so
	// that an unset value is omitted: Tandoor's list/get queries filter out
	// plans whose to_date falls outside the default visibility window, so a
	// zero-value to_date would make the entry invisible (and undeletable).
	FromDate *time.Time `json:"from_date,omitempty"`

	// ToDate is the end date/time of the meal plan entry; nil lets the
	// server default it to from_date.
	ToDate *time.Time `json:"to_date,omitempty"`

	// CreatedBy is the user ID who created the meal plan.
	CreatedBy int `json:"created_by,omitempty"`
}

// UnmarshalJSON custom unmarshals MealPlan to handle meal_type as int or object.
func (m *MealPlan) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID           int              `json:"id,omitempty"`
		Title        string           `json:"title,omitempty"`
		Recipe       *recipe.Overview `json:"recipe,omitempty"`
		RecipeName   string           `json:"recipe_name,omitempty"`
		MealTypeRaw  json.RawMessage  `json:"meal_type,omitempty"`
		MealTypeName string           `json:"meal_type_name,omitempty"`
		Note         string           `json:"note,omitempty"`
		NoteMarkdown string           `json:"note_markdown,omitempty"`
		Servings     float64          `json:"servings,omitempty"`
		Shopping     bool             `json:"shopping,omitempty"`
		AddShopping  bool             `json:"addshopping,omitempty"`
		FromDate     *time.Time       `json:"from_date,omitempty"`
		ToDate       *time.Time       `json:"to_date,omitempty"`
		CreatedBy    int              `json:"created_by,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*m = MealPlan{
		ID:           aux.ID,
		Title:        aux.Title,
		Recipe:       aux.Recipe,
		RecipeName:   aux.RecipeName,
		MealTypeName: aux.MealTypeName,
		Note:         aux.Note,
		NoteMarkdown: aux.NoteMarkdown,
		Servings:     aux.Servings,
		Shopping:     aux.Shopping,
		AddShopping:  aux.AddShopping,
		FromDate:     aux.FromDate,
		ToDate:       aux.ToDate,
		CreatedBy:    aux.CreatedBy,
	}
	// meal_type can be int or object
	if len(aux.MealTypeRaw) > 0 {
		var mtID int
		if err := json.Unmarshal(aux.MealTypeRaw, &mtID); err == nil {
			m.MealTypeID = mtID
		} else {
			var mt MealType
			if err := json.Unmarshal(aux.MealTypeRaw, &mt); err == nil {
				m.MealType = &mt
				m.MealTypeID = mt.ID
			}
		}
	}
	return nil
}

// MealType defines a category of meal (e.g. breakfast, lunch, dinner, snack).
type MealType struct {
	ID        int     `json:"id,omitempty"`
	Name      string  `json:"name"`
	Order     int     `json:"order,omitempty"`
	Time      *string `json:"time,omitempty"`  // JSON time string, e.g. "12:00:00"
	Color     *string `json:"color,omitempty"` // hex color
	CreatedBy int     `json:"created_by,omitempty"`
}

// AutoMealPlanRequest is the request body for the auto-plan endpoint.
type AutoMealPlanRequest struct {
	StartDate   time.Time    `json:"start_date"`
	EndDate     time.Time    `json:"end_date"`
	MealTypeID  int          `json:"meal_type_id"`
	Keywords    []int        `json:"keywords,omitempty"`
	KeywordMode string       `json:"keyword_mode,omitempty"` // "or" or "and"
	Servings    float64      `json:"servings"`
	Shared      []SharedUser `json:"shared,omitempty"`
	AddShopping bool         `json:"addshopping"`
}

// SharedUser is the minimal user reference accepted by the auto-plan
// endpoint's "shared" field (a list of user objects, not bare IDs).
type SharedUser struct {
	ID int `json:"id"`
}
