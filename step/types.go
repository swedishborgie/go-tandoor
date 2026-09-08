// Package step contains types for Tandoor Step resources.
package step

// Step is a cooking step within a recipe, as returned by the standalone
// /api/step/ endpoint.
type Step struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	Instruction          string `json:"instruction"`
	Time                 int    `json:"time"`
	Order                int    `json:"order"`
	ShowAsHeader         bool   `json:"show_as_header"`
	ShowIngredientsTable bool   `json:"show_ingredients_table"`
}
