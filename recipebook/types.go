// Package recipebook contains types for Tandoor Recipe Book resources.
package recipebook

import (
	"github.com/swedishborgie/go-tandoor/recipe"
)

// RecipeBook represents a collection of recipes.
type RecipeBook struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Shared      []recipe.User `json:"shared"`
	CreatedBy   recipe.User   `json:"created_by"`
	Filter      any           `json:"filter"`
	Order       int           `json:"order"`
}

// Entry links a recipe to a recipe book.
type Entry struct {
	ID            int              `json:"id"`
	Book          int              `json:"book"`
	BookContent   *RecipeBook      `json:"book_content"`
	Recipe        int              `json:"recipe"`
	RecipeContent *recipe.Overview `json:"recipe_content"`
}
