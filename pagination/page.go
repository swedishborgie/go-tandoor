// Package pagination provides types for DRF-style paginated API responses.
//
// Tandoor uses Django REST Framework's PageNumberPagination which returns
// responses of the form:
//
//	{
//	    "count": 1234,
//	    "next": "https://api.example.com/api/recipe/?page=2",
//	    "previous": null,
//	    "results": [...]
//	}
//
// This package provides the [Paginated] generic struct to represent such
// responses, and [ListOptions] for common query parameters.
package pagination

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// Paginated is a generic wrapper for a paginated API response.
type Paginated[T any] struct {
	// Count is the total number of items across all pages.
	Count int `json:"count"`

	// Next is the URL for the next page of results, or empty if on the last page.
	Next *string `json:"next,omitempty"`

	// Previous is the URL for the previous page of results, or empty if on the first page.
	Previous *string `json:"previous,omitempty"`

	// Results is the slice of items on the current page.
	Results []T `json:"results"`
}

// HasNext returns true if there are more pages of results.
func (p *Paginated[T]) HasNext() bool {
	return p.Next != nil && *p.Next != ""
}

// HasPrevious returns true if there is a previous page of results.
func (p *Paginated[T]) HasPrevious() bool {
	return p.Previous != nil && *p.Previous != ""
}

// NextPageNumber returns the page number implied by the Next URL, or 0 if
// the URL cannot be parsed.
func (p *Paginated[T]) NextPageNumber() int {
	if !p.HasNext() {
		return 0
	}
	return extractPageNumber(*p.Next)
}

// PreviousPageNumber returns the page number implied by the Previous URL.
func (p *Paginated[T]) PreviousPageNumber() int {
	if !p.HasPrevious() {
		return 0
	}
	return extractPageNumber(*p.Previous)
}

// extractPageNumber parses the "page" query parameter from a URL.
func extractPageNumber(rawURL string) int {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	pageStr := u.Query().Get("page")
	if pageStr == "" {
		return 0
	}
	var page int
	_, err = fmt.Sscanf(pageStr, "%d", &page)
	if err != nil {
		return 0
	}
	return page
}

// ListOptions holds common query parameters for paginated list endpoints.
type ListOptions struct {
	// Page is the page number (1-indexed).
	Page int

	// PageSize controls the number of results per page.
	PageSize int

	// Search is a name search query. Tandoor list endpoints read the
	// `query` query parameter (icontains/trigram name lookup), not `search`.
	Search string

	// OrderBy is the ordering field, e.g. "name" or "-name" for descending.
	OrderBy string

	// Extra allows arbitrary query parameters not covered above.
	Extra map[string]string
}

// Values builds a url.Values from ListOptions.
func (o *ListOptions) Values() url.Values {
	v := url.Values{}
	if o.Page > 0 {
		v.Set("page", fmt.Sprintf("%d", o.Page))
	}
	if o.PageSize > 0 {
		v.Set("page_size", fmt.Sprintf("%d", o.PageSize))
	}
	if o.Search != "" {
		v.Set("query", o.Search)
	}
	if o.OrderBy != "" {
		v.Set("ordering", o.OrderBy)
	}
	for k, val := range o.Extra {
		if val != "" {
			v.Set(k, val)
		}
	}
	return v
}

// QueryString returns the query string portion (without leading '?').
func (o *ListOptions) QueryString() string {
	v := o.Values()
	s := v.Encode()
	s = strings.TrimPrefix(s, "?")
	return s
}

// ListOptionsBuilder provides a fluent builder for ListOptions.
type ListOptionsBuilder struct {
	opts ListOptions
}

// NewListOptions creates a new ListOptionsBuilder.
func NewListOptions() *ListOptionsBuilder {
	return &ListOptionsBuilder{}
}

// Page sets the page number.
func (b *ListOptionsBuilder) Page(n int) *ListOptionsBuilder {
	b.opts.Page = n
	return b
}

// PageSize sets the page size.
func (b *ListOptionsBuilder) PageSize(n int) *ListOptionsBuilder {
	b.opts.PageSize = n
	return b
}

// Query sets the search query (maps to `query` param).
func (b *ListOptionsBuilder) Query(q string) *ListOptionsBuilder {
	b.opts.Search = q
	return b
}

// OrderBy sets the ordering, e.g. "name" or "-name".
func (b *ListOptionsBuilder) OrderBy(field string, asc bool) *ListOptionsBuilder {
	if field == "" {
		return b
	}
	if asc {
		b.opts.OrderBy = field
	} else {
		b.opts.OrderBy = "-" + field
	}
	return b
}

// WithExtra adds an arbitrary query parameter.
func (b *ListOptionsBuilder) WithExtra(key, value string) *ListOptionsBuilder {
	if b.opts.Extra == nil {
		b.opts.Extra = make(map[string]string)
	}
	b.opts.Extra[key] = value
	return b
}

// Build returns the built ListOptions.
func (b *ListOptionsBuilder) Build() ListOptions {
	return b.opts
}

// Iterator provides a simple way to iterate over all pages of results.
//
// Use the client's List method to get the first page, then call Next()
// until it returns false.
type Iterator[T any] struct {
	fetch func() (*Paginated[T], error)
	page  *Paginated[T]
	index int
	err   error
}

// NewIterator creates a new paginated iterator.
//
// fetch should be a closure that calls the client's List method for the
// next page (typically by incrementing the page number).
func NewIterator[T any](fetch func() (*Paginated[T], error)) *Iterator[T] {
	return &Iterator[T]{fetch: fetch}
}

// Next advances to the next item. Returns false when all items have been
// consumed or an error occurred. Use [Iterator.Err] to check for errors.
func (it *Iterator[T]) Next() bool {
	for {
		// Fetch the next page if we're at the start or exhausted the current page.
		if it.page == nil || it.index >= len(it.page.Results) {
			if it.page != nil && !it.page.HasNext() {
				return false
			}
			p, err := it.fetch()
			if err != nil {
				it.err = err
				return false
			}
			it.page = p
			it.index = 0
			continue
		}
		it.index++
		return true
	}
}

// Current returns the current item. Must be called after a successful [Iterator.Next].
func (it *Iterator[T]) Current() T {
	if it.page != nil && it.index > 0 && it.index <= len(it.page.Results) {
		return it.page.Results[it.index-1]
	}
	var zero T
	return zero
}

// Page returns the current page (useful for accessing Count, Next, Previous).
func (it *Iterator[T]) Page() *Paginated[T] {
	return it.page
}

// Err returns the first error encountered, if any.
func (it *Iterator[T]) Err() error {
	return it.err
}

// CollectAll collects all items from a paginated response by repeatedly
// fetching subsequent pages using nextFn. The first page is provided
// explicitly so the caller can control the initial request.
// CollectAll returns a flat slice of all items across all pages.
func CollectAll[T any](_ context.Context, first *Paginated[T], nextFn func(page int) (*Paginated[T], error)) ([]T, error) {
	if first == nil {
		return nil, fmt.Errorf("pagination: first page is nil")
	}
	var all []T
	all = append(all, first.Results...)

	current := first
	for current.HasNext() {
		pageNum := current.NextPageNumber()
		if pageNum == 0 {
			// Fallback to incrementing based on results already collected
			pageNum = 2
		}
		next, err := nextFn(pageNum)
		if err != nil {
			return all, err
		}
		if next == nil {
			break
		}
		all = append(all, next.Results...)
		current = next
	}
	return all, nil
}
