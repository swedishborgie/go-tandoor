// Package step provides typed list options.
package step

import (
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for step list endpoints.
type ListOptions struct {
	pagination.ListOptions
	RecipeID int
}

// Values builds url.Values.
func (o ListOptions) Values() url.Values {
	v := o.ListOptions.Values()
	if o.RecipeID != 0 {
		v.Set("recipe", strconv.Itoa(o.RecipeID))
	}
	return v
}

// ToPaginationOptions converts to pagination.ListOptions.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	return &base
}
