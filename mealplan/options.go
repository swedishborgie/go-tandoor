// Package mealplan provides typed list options.
package mealplan

import (
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for mealplan list endpoints.
type ListOptions struct {
	pagination.ListOptions
	SpaceID  int
	UserID   int
	DateFrom string
	DateTo   string
}

// Values builds url.Values.
func (o ListOptions) Values() url.Values {
	v := o.ListOptions.Values()
	if o.SpaceID != 0 {
		v.Set("space", strconv.Itoa(o.SpaceID))
	}
	if o.UserID != 0 {
		v.Set("user", strconv.Itoa(o.UserID))
	}
	if o.DateFrom != "" {
		v.Set("date_from", o.DateFrom)
	}
	if o.DateTo != "" {
		v.Set("date_to", o.DateTo)
	}
	return v
}

// QueryString returns the query string for the options, including the
// typed filters (shadowing the promoted pagination implementation, which
// would drop SpaceID/UserID/DateFrom/DateTo).
func (o ListOptions) QueryString() string {
	return o.Values().Encode()
}

// ToPaginationOptions converts to pagination.ListOptions.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	extra := make(map[string]string)
	if o.SpaceID != 0 {
		extra["space"] = strconv.Itoa(o.SpaceID)
	}
	if o.UserID != 0 {
		extra["user"] = strconv.Itoa(o.UserID)
	}
	if o.DateFrom != "" {
		extra["date_from"] = o.DateFrom
	}
	if o.DateTo != "" {
		extra["date_to"] = o.DateTo
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
