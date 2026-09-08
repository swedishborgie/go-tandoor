package keyword

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Keyword API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new KeywordService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of keywords.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Keyword], error) {
	path := "api/keyword/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Keyword]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAll returns all keywords matching opts by iterating over all pages.
func (s *Service) ListAll(ctx context.Context, opts *ListOptions) ([]Keyword, error) {
	if opts == nil {
		opts = &ListOptions{}
	}
	firstOpts := *opts
	if firstOpts.Page == 0 {
		firstOpts.Page = 1
	}
	if firstOpts.PageSize == 0 {
		firstOpts.PageSize = 100
	}
	firstPage, err := s.List(ctx, &firstOpts)
	if err != nil {
		return nil, err
	}
	nextFn := func(page int) (*pagination.Paginated[Keyword], error) {
		optsCopy := *opts
		optsCopy.Page = page
		if optsCopy.PageSize == 0 {
			optsCopy.PageSize = 100
		}
		return s.List(ctx, &optsCopy)
	}
	return pagination.CollectAll(ctx, firstPage, nextFn)
}

// Get retrieves a keyword by ID.
func (s *Service) Get(ctx context.Context, id int) (*Keyword, error) {
	var k Keyword
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/keyword/%d/", id), nil, &k); err != nil {
		return nil, err
	}
	return &k, nil
}

// Create creates a new keyword.
func (s *Service) Create(ctx context.Context, k *Keyword) (*Keyword, error) {
	var result Keyword
	if err := s.exec.DoJSON(ctx, "POST", "api/keyword/", k, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a keyword.
func (s *Service) Update(ctx context.Context, k *Keyword) (*Keyword, error) {
	var result Keyword
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/keyword/%d/", k.ID), k, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a keyword.
func (s *Service) Patch(ctx context.Context, k *Keyword) (*Keyword, error) {
	var result Keyword
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/keyword/%d/", k.ID), k, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a keyword by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/keyword/%d/", id), nil, nil)
}

// Merge merges one keyword into another.
func (s *Service) Merge(ctx context.Context, sourceID, targetID int) (*Keyword, error) {
	var result Keyword
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/keyword/%d/merge/%d/", sourceID, targetID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Move moves a keyword to a different parent.
func (s *Service) Move(ctx context.Context, id, parentID int) (*Keyword, error) {
	var result Keyword
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/keyword/%d/move/%d/", id, parentID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Cascading returns objects that would be cascade-deleted with this keyword.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Keyword], error) {
	path := fmt.Sprintf("api/keyword/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Keyword]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this keyword is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Keyword], error) {
	path := fmt.Sprintf("api/keyword/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Keyword]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this keyword from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Keyword], error) {
	path := fmt.Sprintf("api/keyword/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Keyword]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
