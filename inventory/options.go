// Package inventory provides typed list options.
package inventory

import (
	"github.com/swedishborgie/go-tandoor/pagination"
	"net/url"
)

// LocationListOptions provides typed list options.
type LocationListOptions struct{ pagination.ListOptions }

// Values builds url.Values.
func (o LocationListOptions) Values() url.Values { return o.ListOptions.Values() }
