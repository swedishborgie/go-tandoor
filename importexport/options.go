// Package importexport provides typed list options.
package importexport

import (
	"github.com/swedishborgie/go-tandoor/pagination"
	"net/url"
)

// RecipeImportListOptions provides typed list options.
type RecipeImportListOptions struct{ pagination.ListOptions }

// ImportLogListOptions provides typed list options.
type ImportLogListOptions struct{ pagination.ListOptions }

// ExportLogListOptions provides typed list options.
type ExportLogListOptions struct{ pagination.ListOptions }

// BookmarkletImportListOptions provides typed list options.
type BookmarkletImportListOptions struct{ pagination.ListOptions }

// SyncListOptions provides typed list options.
type SyncListOptions struct{ pagination.ListOptions }

// SyncLogListOptions provides typed list options.
type SyncLogListOptions struct{ pagination.ListOptions }

// Values builds url.Values.
func (o RecipeImportListOptions) Values() url.Values { return o.ListOptions.Values() }

// Values builds url.Values.
func (o ImportLogListOptions) Values() url.Values { return o.ListOptions.Values() }

// Values builds url.Values.
func (o ExportLogListOptions) Values() url.Values { return o.ListOptions.Values() }

// Values builds url.Values.
func (o BookmarkletImportListOptions) Values() url.Values { return o.ListOptions.Values() }

// Values builds url.Values.
func (o SyncListOptions) Values() url.Values { return o.ListOptions.Values() }

// Values builds url.Values.
func (o SyncLogListOptions) Values() url.Values { return o.ListOptions.Values() }
