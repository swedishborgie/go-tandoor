// Package keyword provides typed list options.
package keyword

import (
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for keyword list endpoints.
type ListOptions struct {
	pagination.ListOptions
	RecipeBookID int
}

// Values builds url.Values.
func (o ListOptions) Values() url.Values {
	v := o.ListOptions.Values()
	if o.RecipeBookID != 0 {
		v.Set("recipe_book", strconv.Itoa(o.RecipeBookID))
	}
	return v
}

// ToPaginationOptions converts to pagination.ListOptions.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	if o.RecipeBookID != 0 {
		if base.Extra == nil {
			base.Extra = make(map[string]string)
		}
		base.Extra["recipe_book"] = strconv.Itoa(o.RecipeBookID)
	}
	return &base
}
