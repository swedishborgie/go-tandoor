// Package supermarket contains types for Tandoor Supermarket resources.
package supermarket

// Supermarket represents a store where food can be purchased.
//
// Endpoints: GET/POST api/supermarket/ GET/PUT/PATCH/DELETE api/supermarket/<id>/
type Supermarket struct {
	ID                    int                `json:"id"`
	Name                  string             `json:"name"`
	Description           string             `json:"description"`
	ShoppingLists         []int              `json:"shopping_lists"`
	CategoryToSupermarket []CategoryRelation `json:"category_to_supermarket"`
	OpenDataSlug          string             `json:"open_data_slug"`
}

// Category represents a category (e.g., "Dairy", "Produce").
//
// Endpoints: GET/POST api/supermarket-category/ GET/PUT/PATCH/DELETE api/supermarket-category/<id>/
// Actions: merge, cascading/nulling/protecting delete
type Category struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	OpenDataSlug string `json:"open_data_slug"`
}

// CategoryRelation links a category to a supermarket.
//
// Endpoints: GET/POST api/supermarket-category-relation/ GET/PUT/PATCH/DELETE api/supermarket-category-relation/<id>/
type CategoryRelation struct {
	ID          int       `json:"id"`
	Category    *Category `json:"category"`
	Supermarket *int      `json:"supermarket"`
	Order       int       `json:"order"`
}
