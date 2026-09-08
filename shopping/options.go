// Package shopping provides typed list options.
package shopping

import (
	"github.com/swedishborgie/go-tandoor/pagination"
	"net/url"
)

// ListListOptions provides typed list options.
type ListListOptions struct{ pagination.ListOptions }

// ListEntryListOptions provides typed list options.
type ListEntryListOptions struct{ pagination.ListOptions }

// ListRecipeListOptions provides typed list options.
type ListRecipeListOptions struct{ pagination.ListOptions }

// Values builds url.Values.
func (o ListListOptions) Values() url.Values { return o.ListOptions.Values() }

// Values builds url.Values.
func (o ListEntryListOptions) Values() url.Values { return o.ListOptions.Values() }

// Values builds url.Values.
func (o ListRecipeListOptions) Values() url.Values { return o.ListOptions.Values() }
