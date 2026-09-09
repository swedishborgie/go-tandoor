// Package recipe contains types for Tandoor Recipe resources.
package recipe

import (
	"time"

	"github.com/swedishborgie/go-tandoor/idref"
)

// Recipe is the full recipe model.
type Recipe struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`

	Keywords []Keyword `json:"keywords,omitempty"`
	Steps    []Step    `json:"steps"`

	WorkingTime int `json:"working_time,omitempty"`
	WaitingTime int `json:"waiting_time,omitempty"`

	CreatedBy User      `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	SourceURL              string `json:"source_url,omitempty"`
	Internal               bool   `json:"internal"`
	Private                bool   `json:"private"`
	ShowIngredientOverview bool   `json:"show_ingredient_overview"`

	Nutrition      *NutritionInformation `json:"nutrition,omitempty"`
	FoodProperties any                   `json:"food_properties,omitempty"`

	Servings     int       `json:"servings,omitempty"`
	ServingsText string    `json:"servings_text,omitempty"`
	Diameter     int       `json:"diameter,omitempty"`
	DiameterText string    `json:"diameter_text,omitempty"`
	Rating       float64   `json:"rating,omitempty"`
	LastCooked   time.Time `json:"last_cooked,omitempty"`
	New          bool      `json:"new,omitempty"`
	Recent       string    `json:"recent,omitempty"`

	FilePath string `json:"file_path,omitempty"`

	// Shopping-related fields
	ShoppingListRecipes []ShoppingListRecipe `json:"shopping_list_recipe,omitempty"`
	ShoppingServings    []int                `json:"shopping_servings,omitempty"`
}

// Simple is a minimal recipe reference (id, name, url).
type Simple struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Flat is a flat list item (id, name, image).
type Flat struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// Overview is the list-card view of a recipe.
type Overview struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Image        string         `json:"image"`
	Keywords     []KeywordLabel `json:"keywords,omitempty"`
	WorkingTime  int            `json:"working_time"`
	WaitingTime  int            `json:"waiting_time"`
	CreatedBy    User           `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Internal     bool           `json:"internal"`
	Private      bool           `json:"private"`
	Servings     int            `json:"servings"`
	ServingsText string         `json:"servings_text"`
	Rating       float64        `json:"rating"`
	LastCooked   time.Time      `json:"last_cooked"`
	New          bool           `json:"new"`
	Recent       string         `json:"recent"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full object (mirrors Tandoor's writable nested handling).
func (o *Overview) MarshalJSON() ([]byte, error) {
	type plain Overview
	return idref.MarshalJSON(o.ID, o.Name, plain(*o))
}

// UnmarshalJSON accepts either a bare integer or the full object.
func (o *Overview) UnmarshalJSON(data []byte) error {
	type plain Overview
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*o = Overview(p)
	return nil
}

// BatchUpdate performs batch operations on multiple recipes.
type BatchUpdate struct {
	Recipes           []int `json:"recipes"`
	KeywordsAdd       []int `json:"keywords_add,omitempty"`
	KeywordsRemove    []int `json:"keywords_remove,omitempty"`
	KeywordsSet       []int `json:"keywords_set,omitempty"`
	KeywordsRemoveAll bool  `json:"keywords_remove_all"`
	WorkingTime       *int  `json:"working_time,omitempty"`
	WaitingTime       *int  `json:"waiting_time,omitempty"`
}

// Image updates a recipe's image.
type Image struct {
	Image    string `json:"image,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// ShoppingUpdate adds a recipe to a shopping list.
type ShoppingUpdate struct {
	ID           int   `json:"id"`
	ListRecipeID *int  `json:"list_recipe,omitempty"`
	Ingredients  []int `json:"ingredients"`
	Servings     *int  `json:"servings,omitempty"`
}

// Keyword is a keyword/label for recipes.
type Keyword struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Parent      int       `json:"parent"`
	NumChild    int       `json:"numchild"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FullName    string    `json:"full_name"`
}

// KeywordLabel is a minimal keyword reference (id, label).
type KeywordLabel struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

// Step is a cooking step within a recipe.
type Step struct {
	ID                   int          `json:"id"`
	Name                 string       `json:"name"`
	Instruction          string       `json:"instruction"`
	Ingredients          []Ingredient `json:"ingredients"`
	InstructionsMarkdown string       `json:"instructions_markdown"`
	Time                 int          `json:"time"`
	Order                int          `json:"order"`
	ShowAsHeader         bool         `json:"show_as_header"`
	StepRecipe           *int         `json:"step_recipe"`
	StepRecipeData       any          `json:"step_recipe_data"`
	NumRecipe            int          `json:"numrecipe"`
	ShowIngredientsTable bool         `json:"show_ingredients_table"`
}

// Ingredient is an ingredient within a step.
type Ingredient struct {
	ID            int     `json:"id"`
	Food          *Food   `json:"food"`
	Unit          *Unit   `json:"unit"`
	Amount        float64 `json:"amount"`
	Conversions   []any   `json:"conversions"`
	Note          string  `json:"note"`
	Order         int     `json:"order"`
	IsHeader      bool    `json:"is_header"`
	NoAmount      bool    `json:"no_amount"`
	OriginalText  string  `json:"original_text"`
	UsedInRecipes []any   `json:"used_in_recipes"`
	Checked       bool    `json:"checked"`
}

// IngredientSimple is the ingredient reference without computed fields.
type IngredientSimple struct {
	ID           int         `json:"id"`
	Food         *FoodSimple `json:"food"`
	Unit         *Unit       `json:"unit"`
	Amount       float64     `json:"amount"`
	Note         string      `json:"note"`
	Order        int         `json:"order"`
	IsHeader     bool        `json:"is_header"`
	NoAmount     bool        `json:"no_amount"`
	OriginalText string      `json:"original_text"`
	Checked      bool        `json:"checked"`
}

// FoodSimple is a minimal food reference.
type FoodSimple struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PluralName string `json:"plural_name,omitempty"`
}

// Unit is a measurement unit.
type Unit struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PluralName   string `json:"plural_name,omitempty"`
	Description  string `json:"description"`
	BaseUnit     string `json:"base_unit,omitempty"`
	OpenDataSlug string `json:"open_data_slug,omitempty"`
}

// User is a minimal user reference.
type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	IsStaff     bool   `json:"is_staff"`
	IsSuperuser bool   `json:"is_superuser"`
	IsActive    bool   `json:"is_active"`
}

// MarshalJSON marshals as a bare integer when only an ID is set, otherwise
// as the full user object (mirrors Tandoor's writable nested handling).
func (u *User) MarshalJSON() ([]byte, error) {
	type plain User
	return idref.MarshalJSON(u.ID, u.Username, plain(*u))
}

// UnmarshalJSON accepts either a bare integer or the full user object.
func (u *User) UnmarshalJSON(data []byte) error {
	type plain User
	var p plain
	if err := idref.UnmarshalJSON(data, &p.ID, &p); err != nil {
		return err
	}
	*u = User(p)
	return nil
}

// NutritionInformation holds nutrition data for a recipe.
type NutritionInformation struct {
	ID            int     `json:"id"`
	Carbohydrates float64 `json:"carbohydrates"`
	Fats          float64 `json:"fats"`
	Proteins      float64 `json:"proteins"`
	Calories      float64 `json:"calories"`
	Source        string  `json:"source,omitempty"`
}

// Food is the full food model (used in nested relations).
type Food struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	PluralName       string  `json:"plural_name"`
	Description      string  `json:"description"`
	Shopping         string  `json:"shopping"`
	Recipe           *Simple `json:"recipe"`
	URL              string  `json:"url"`
	FDCID            *int    `json:"fdc_id"`
	IgnoreShopping   bool    `json:"ignore_shopping"`
	OpenDataSlug     string  `json:"open_data_slug"`
	Parent           int     `json:"parent"`
	NumChild         int     `json:"numchild"`
	FullName         string  `json:"full_name"`
	SubstituteOnHand bool    `json:"substitute_onhand"`
}

// ShoppingListRecipe references a recipe added to a shopping list.
type ShoppingListRecipe struct {
	ID       int `json:"id"`
	Quantity int `json:"quantity"`
}
