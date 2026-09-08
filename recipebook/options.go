// Package recipebook provides typed list options.
package recipebook

import (
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for recipe book list endpoints.
type ListOptions struct {
	pagination.ListOptions
	SpaceID int
}

// Values builds url.Values.
func (o ListOptions) Values() url.Values {
	v := o.ListOptions.Values()
	if o.SpaceID != 0 {
		v.Set("space", strconv.Itoa(o.SpaceID))
	}
	return v
}

// ToPaginationOptions converts to pagination.ListOptions.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	if o.SpaceID != 0 {
		if base.Extra == nil {
			base.Extra = make(map[string]string)
		}
		base.Extra["space"] = strconv.Itoa(o.SpaceID)
	}
	return &base
}
