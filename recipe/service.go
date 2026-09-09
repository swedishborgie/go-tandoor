package recipe

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Recipe API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new RecipeService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of recipes.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Recipe], error) {
	path := "api/recipe/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Recipe]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAll returns all recipes matching opts by iterating over all pages.
func (s *Service) ListAll(ctx context.Context, opts *ListOptions) ([]Recipe, error) {
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
	nextFn := func(page int) (*pagination.Paginated[Recipe], error) {
		optsCopy := *opts
		optsCopy.Page = page
		if optsCopy.PageSize == 0 {
			optsCopy.PageSize = 100
		}
		return s.List(ctx, &optsCopy)
	}
	return pagination.CollectAll(ctx, firstPage, nextFn)
}

// Flat returns a list of flat recipes (id, name, image). The endpoint
// responds with a flat array, not a paginated envelope.
func (s *Service) Flat(ctx context.Context, opts *ListOptions) ([]Flat, error) {
	path := "api/recipe/flat/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var result []Flat
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Get retrieves a single recipe by ID.
func (s *Service) Get(ctx context.Context, id int) (*Recipe, error) {
	var r Recipe
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe/%d/", id), nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Create creates a new recipe.
func (s *Service) Create(ctx context.Context, r *Recipe) (*Recipe, error) {
	var result Recipe
	if err := s.exec.DoJSON(ctx, "POST", "api/recipe/", r, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a recipe.
func (s *Service) Update(ctx context.Context, r *Recipe) (*Recipe, error) {
	var result Recipe
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/recipe/%d/", r.ID), r, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a recipe.
func (s *Service) Patch(ctx context.Context, r *Recipe) (*Recipe, error) {
	var result Recipe
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/recipe/%d/", r.ID), r, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a recipe by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe/%d/", id), nil, nil)
}

// Related returns recipes related to the given recipe.
func (s *Service) Related(ctx context.Context, id int, opts *ListOptions) (*pagination.Paginated[Recipe], error) {
	path := fmt.Sprintf("api/recipe/%d/related/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Recipe]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// BatchUpdate updates keywords and times for multiple recipes at once.
func (s *Service) BatchUpdate(ctx context.Context, update *BatchUpdate) (*pagination.Paginated[Recipe], error) {
	var page pagination.Paginated[Recipe]
	if err := s.exec.DoJSON(ctx, "PUT", "api/recipe/batch_update/", update, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// UploadImage updates a recipe's image.
func (s *Service) UploadImage(ctx context.Context, id int, img *Image) (*Recipe, error) {
	var result Recipe
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/recipe/%d/image/", id), img, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddToShopping adds a recipe to a shopping list.
func (s *Service) AddToShopping(ctx context.Context, id int, update *ShoppingUpdate) error {
	update.ID = id
	return s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/recipe/%d/shopping/", id), update, nil)
}

// Cascading returns objects that would be cascade-deleted with this recipe.
func (s *Service) Cascading(ctx context.Context, id int, opts *ListOptions) (*pagination.Paginated[Recipe], error) {
	path := fmt.Sprintf("api/recipe/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Recipe]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this recipe is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *ListOptions) (*pagination.Paginated[Recipe], error) {
	path := fmt.Sprintf("api/recipe/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Recipe]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this recipe from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *ListOptions) (*pagination.Paginated[Recipe], error) {
	path := fmt.Sprintf("api/recipe/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Recipe]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// AiProperties triggers AI to generate properties for a recipe.
func (s *Service) AiProperties(ctx context.Context, id int) (*Recipe, error) {
	var result Recipe
	if err := s.exec.DoJSON(ctx, "POST", fmt.Sprintf("api/recipe/%d/aiproperties/", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteExternal removes the external file from a recipe.
func (s *Service) DeleteExternal(ctx context.Context, id int) (*Recipe, error) {
	var result Recipe
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/recipe/%d/delete_external/", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
