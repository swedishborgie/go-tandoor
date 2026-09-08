package supermarket

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Supermarket API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new SupermarketService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of supermarkets.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Supermarket], error) {
	path := "api/supermarket/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Supermarket]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single supermarket by ID.
func (s *Service) Get(ctx context.Context, id int) (*Supermarket, error) {
	var sm Supermarket
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/supermarket/%d/", id), nil, &sm); err != nil {
		return nil, err
	}
	return &sm, nil
}

// Create creates a new supermarket.
func (s *Service) Create(ctx context.Context, sm *Supermarket) (*Supermarket, error) {
	var result Supermarket
	if err := s.exec.DoJSON(ctx, "POST", "api/supermarket/", sm, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a supermarket.
func (s *Service) Update(ctx context.Context, sm *Supermarket) (*Supermarket, error) {
	var result Supermarket
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/supermarket/%d/", sm.ID), sm, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a supermarket.
func (s *Service) Patch(ctx context.Context, sm *Supermarket) (*Supermarket, error) {
	var result Supermarket
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/supermarket/%d/", sm.ID), sm, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a supermarket.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/supermarket/%d/", id), nil, nil)
}

// CategoryService provides access to Supermarket Category API endpoints.
type CategoryService struct {
	exec executor.Executor
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(e executor.Executor) *CategoryService {
	return &CategoryService{exec: e}
}

// List returns a paginated list of supermarket categories.
func (s *CategoryService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[Category], error) {
	path := "api/supermarket-category/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Category]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single supermarket category by ID.
func (s *CategoryService) Get(ctx context.Context, id int) (*Category, error) {
	var cat Category
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/supermarket-category/%d/", id), nil, &cat); err != nil {
		return nil, err
	}
	return &cat, nil
}

// Create creates a new supermarket category.
func (s *CategoryService) Create(ctx context.Context, cat *Category) (*Category, error) {
	var result Category
	if err := s.exec.DoJSON(ctx, "POST", "api/supermarket-category/", cat, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a supermarket category.
func (s *CategoryService) Update(ctx context.Context, cat *Category) (*Category, error) {
	var result Category
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/supermarket-category/%d/", cat.ID), cat, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a supermarket category.
func (s *CategoryService) Patch(ctx context.Context, cat *Category) (*Category, error) {
	var result Category
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/supermarket-category/%d/", cat.ID), cat, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a supermarket category.
func (s *CategoryService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/supermarket-category/%d/", id), nil, nil)
}

// Merge merges another supermarket category into this one.
func (s *CategoryService) Merge(ctx context.Context, id, other int) (*Category, error) {
	var result Category
	if err := s.exec.DoJSON(ctx, "POST", fmt.Sprintf("api/supermarket-category/%d/merge/?other=%d", id, other), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CategoryRelationService provides access to Supermarket Category Relation API endpoints.
type CategoryRelationService struct {
	exec executor.Executor
}

// NewCategoryRelationService creates a new CategoryRelationService.
func NewCategoryRelationService(e executor.Executor) *CategoryRelationService {
	return &CategoryRelationService{exec: e}
}

// List returns a paginated list of supermarket category relations.
func (s *CategoryRelationService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[CategoryRelation], error) {
	path := "api/supermarket-category-relation/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[CategoryRelation]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single supermarket category relation by ID.
func (s *CategoryRelationService) Get(ctx context.Context, id int) (*CategoryRelation, error) {
	var rel CategoryRelation
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/supermarket-category-relation/%d/", id), nil, &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// Create creates a new supermarket category relation.
func (s *CategoryRelationService) Create(ctx context.Context, rel *CategoryRelation) (*CategoryRelation, error) {
	var result CategoryRelation
	if err := s.exec.DoJSON(ctx, "POST", "api/supermarket-category-relation/", rel, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a supermarket category relation.
func (s *CategoryRelationService) Update(ctx context.Context, rel *CategoryRelation) (*CategoryRelation, error) {
	var result CategoryRelation
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/supermarket-category-relation/%d/", rel.ID), rel, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a supermarket category relation.
func (s *CategoryRelationService) Patch(ctx context.Context, rel *CategoryRelation) (*CategoryRelation, error) {
	var result CategoryRelation
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/supermarket-category-relation/%d/", rel.ID), rel, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a supermarket category relation.
func (s *CategoryRelationService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/supermarket-category-relation/%d/", id), nil, nil)
}
