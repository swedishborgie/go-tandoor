package recipebook

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Recipe Book API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new RecipeBookService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of recipe books.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[RecipeBook], error) {
	path := "api/recipe-book/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[RecipeBook]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a recipe book by ID.
func (s *Service) Get(ctx context.Context, id int) (*RecipeBook, error) {
	var rb RecipeBook
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe-book/%d/", id), nil, &rb); err != nil {
		return nil, err
	}
	return &rb, nil
}

// Create creates a new recipe book.
func (s *Service) Create(ctx context.Context, rb *RecipeBook) (*RecipeBook, error) {
	var result RecipeBook
	if err := s.exec.DoJSON(ctx, "POST", "api/recipe-book/", rb, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a recipe book.
func (s *Service) Update(ctx context.Context, rb *RecipeBook) (*RecipeBook, error) {
	var result RecipeBook
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/recipe-book/%d/", rb.ID), rb, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a recipe book.
func (s *Service) Patch(ctx context.Context, rb *RecipeBook) (*RecipeBook, error) {
	var result RecipeBook
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/recipe-book/%d/", rb.ID), rb, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a recipe book by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe-book/%d/", id), nil, nil)
}

// Cascading returns objects that would be cascade-deleted with this recipe book.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[RecipeBook], error) {
	path := fmt.Sprintf("api/recipe-book/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[RecipeBook]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this recipe book is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[RecipeBook], error) {
	path := fmt.Sprintf("api/recipe-book/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[RecipeBook]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this recipe book from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[RecipeBook], error) {
	path := fmt.Sprintf("api/recipe-book/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[RecipeBook]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// EntryService provides access to Recipe Book Entry API endpoints.
type EntryService struct {
	exec executor.Executor
}

// NewEntryService creates a new EntryService.
func NewEntryService(e executor.Executor) *EntryService {
	return &EntryService{exec: e}
}

// List returns a paginated list of recipe book entries.
func (s *EntryService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[Entry], error) {
	path := "api/recipe-book-entry/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Entry]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a recipe book entry by ID.
func (s *EntryService) Get(ctx context.Context, id int) (*Entry, error) {
	var rbe Entry
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe-book-entry/%d/", id), nil, &rbe); err != nil {
		return nil, err
	}
	return &rbe, nil
}

// Create creates a new recipe book entry.
func (s *EntryService) Create(ctx context.Context, rbe *Entry) (*Entry, error) {
	var result Entry
	if err := s.exec.DoJSON(ctx, "POST", "api/recipe-book-entry/", rbe, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a recipe book entry.
func (s *EntryService) Update(ctx context.Context, rbe *Entry) (*Entry, error) {
	var result Entry
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/recipe-book-entry/%d/", rbe.ID), rbe, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a recipe book entry.
func (s *EntryService) Patch(ctx context.Context, rbe *Entry) (*Entry, error) {
	var result Entry
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/recipe-book-entry/%d/", rbe.ID), rbe, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a recipe book entry by ID.
func (s *EntryService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe-book-entry/%d/", id), nil, nil)
}
