// Package property provides typed list options.
package property

import (
	"github.com/swedishborgie/go-tandoor/pagination"
	"net/url"
)

// TypeListOptions provides typed list options.
type TypeListOptions struct{ pagination.ListOptions }

// Values builds url.Values.
func (o TypeListOptions) Values() url.Values { return o.ListOptions.Values() }
