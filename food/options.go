// Package food provides typed list options.
package food

import (
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for food list endpoints.
type ListOptions struct {
	pagination.ListOptions
	CategoryID int
	UnitID     int
	// Add more as needed
}

// Values builds url.Values from typed options.
func (o ListOptions) Values() url.Values {
	v := o.ListOptions.Values()
	if o.CategoryID != 0 {
		v.Set("category", strconv.Itoa(o.CategoryID))
	}
	if o.UnitID != 0 {
		v.Set("unit", strconv.Itoa(o.UnitID))
	}
	return v
}

// QueryString returns the query string for the options, including the
// typed filters (shadowing the promoted pagination implementation, which
// would drop CategoryID/UnitID).
func (o ListOptions) QueryString() string {
	return o.Values().Encode()
}

// ToPaginationOptions converts to pagination.ListOptions with Extra.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	extra := make(map[string]string)
	if o.CategoryID != 0 {
		extra["category"] = strconv.Itoa(o.CategoryID)
	}
	if o.UnitID != 0 {
		extra["unit"] = strconv.Itoa(o.UnitID)
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
