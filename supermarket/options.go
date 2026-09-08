// Package supermarket provides typed list options.
package supermarket

import (
	"net/url"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for supermarket list endpoints.
type ListOptions struct {
	pagination.ListOptions
}

// Values builds url.Values.
func (o ListOptions) Values() url.Values {
	return o.ListOptions.Values()
}

// ToPaginationOptions converts to pagination.ListOptions.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	return &base
}
