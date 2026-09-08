package ingredient

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Ingredient API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new IngredientService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of ingredients.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Ingredient], error) {
	path := "api/ingredient/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Ingredient]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves an ingredient by ID.
func (s *Service) Get(ctx context.Context, id int) (*Ingredient, error) {
	var i Ingredient
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/ingredient/%d/", id), nil, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// Create creates a new ingredient.
func (s *Service) Create(ctx context.Context, i *Ingredient) (*Ingredient, error) {
	var result Ingredient
	if err := s.exec.DoJSON(ctx, "POST", "api/ingredient/", i, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of an ingredient.
func (s *Service) Update(ctx context.Context, i *Ingredient) (*Ingredient, error) {
	var result Ingredient
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/ingredient/%d/", i.ID), i, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of an ingredient.
func (s *Service) Patch(ctx context.Context, i *Ingredient) (*Ingredient, error) {
	var result Ingredient
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/ingredient/%d/", i.ID), i, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes an ingredient by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/ingredient/%d/", id), nil, nil)
}

// Cascading returns objects that would be cascade-deleted with this ingredient.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Ingredient], error) {
	path := fmt.Sprintf("api/ingredient/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Ingredient]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this ingredient is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Ingredient], error) {
	path := fmt.Sprintf("api/ingredient/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Ingredient]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this ingredient from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Ingredient], error) {
	path := fmt.Sprintf("api/ingredient/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Ingredient]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
