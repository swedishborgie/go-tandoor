// Package misc provides typed list options.
package misc

import (
	"net/url"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// CookLogListOptions provides typed filtering for cook log list endpoints.
type CookLogListOptions struct {
	pagination.ListOptions
}

// Values builds url.Values.
func (o CookLogListOptions) Values() url.Values {
	return o.ListOptions.Values()
}

// ToPaginationOptions converts to pagination.ListOptions.
func (o CookLogListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	return &base
}
