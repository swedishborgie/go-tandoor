package food

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Food API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new FoodService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of foods.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Food], error) {
	path := "api/food/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Food]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAll returns all foods matching opts by iterating over all pages.
func (s *Service) ListAll(ctx context.Context, opts *ListOptions) ([]Food, error) {
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
	nextFn := func(page int) (*pagination.Paginated[Food], error) {
		optsCopy := *opts
		optsCopy.Page = page
		if optsCopy.PageSize == 0 {
			optsCopy.PageSize = 100
		}
		return s.List(ctx, &optsCopy)
	}
	return pagination.CollectAll(ctx, firstPage, nextFn)
}

// Get retrieves a food by ID.
func (s *Service) Get(ctx context.Context, id int) (*Food, error) {
	var f Food
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/food/%d/", id), nil, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// Create creates a new food.
func (s *Service) Create(ctx context.Context, f *Food) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "POST", "api/food/", f, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a food.
func (s *Service) Update(ctx context.Context, f *Food) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/food/%d/", f.ID), f, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a food.
func (s *Service) Patch(ctx context.Context, f *Food) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/food/%d/", f.ID), f, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a food by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/food/%d/", id), nil, nil)
}

// Merge merges one food into another.
func (s *Service) Merge(ctx context.Context, sourceID, targetID int) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/food/%d/merge/%d/", sourceID, targetID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Move moves a food to a different parent category.
func (s *Service) Move(ctx context.Context, id, parentID int) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/food/%d/move/%d/", id, parentID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchUpdate updates multiple foods at once.
func (s *Service) BatchUpdate(ctx context.Context, update *BatchUpdate) (*pagination.Paginated[Food], error) {
	var page pagination.Paginated[Food]
	if err := s.exec.DoJSON(ctx, "PUT", "api/food/batch_update/", update, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// UpdateShopping updates the shopping behavior for a food.
func (s *Service) UpdateShopping(ctx context.Context, id int, update *ShoppingUpdate) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/food/%d/shopping/", id), update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FdcImport imports FDC data for a food.
func (s *Service) FdcImport(ctx context.Context, id int) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "POST", fmt.Sprintf("api/food/%d/fdc/", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AiProperties triggers AI to generate properties for a food.
func (s *Service) AiProperties(ctx context.Context, id int) (*Food, error) {
	var result Food
	if err := s.exec.DoJSON(ctx, "POST", fmt.Sprintf("api/food/%d/aiproperties/", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Cascading returns objects that would be cascade-deleted with this food.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Food], error) {
	path := fmt.Sprintf("api/food/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Food]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this food is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Food], error) {
	path := fmt.Sprintf("api/food/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Food]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this food from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Food], error) {
	path := fmt.Sprintf("api/food/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Food]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
