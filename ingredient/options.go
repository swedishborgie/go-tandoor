// Package ingredient provides typed list options.
package ingredient

import (
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for ingredient list endpoints.
type ListOptions struct {
	pagination.ListOptions
	FoodID   int
	RecipeID int
	SpaceID  int
}

// Values builds url.Values.
func (o ListOptions) Values() url.Values {
	v := o.ListOptions.Values()
	if o.FoodID != 0 {
		v.Set("food", strconv.Itoa(o.FoodID))
	}
	if o.RecipeID != 0 {
		v.Set("recipe", strconv.Itoa(o.RecipeID))
	}
	if o.SpaceID != 0 {
		v.Set("space", strconv.Itoa(o.SpaceID))
	}
	return v
}

// ToPaginationOptions converts to pagination.ListOptions.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	extra := make(map[string]string)
	if o.FoodID != 0 {
		extra["food"] = strconv.Itoa(o.FoodID)
	}
	if o.RecipeID != 0 {
		extra["recipe"] = strconv.Itoa(o.RecipeID)
	}
	if o.SpaceID != 0 {
		extra["space"] = strconv.Itoa(o.SpaceID)
	}
	if base.Extra == nil {
		base.Extra = extra
	} else {
		for k, v := range extra {
			base.Extra[k] = v
		}
	}
	return &base
}
