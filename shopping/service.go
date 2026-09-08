package shopping

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// ModelReference is a generic model reference used by cascading/nulling/protecting endpoints.
type ModelReference struct {
	ID    int    `json:"id"`
	Model string `json:"model"`
	Name  string `json:"name"`
}

// ListService provides access to Shopping List API endpoints.
type ListService struct {
	exec executor.Executor
}

// NewListService creates a new ListService.
func NewListService(e executor.Executor) *ListService {
	return &ListService{exec: e}
}

// List returns a paginated list of shopping lists.
func (s *ListService) List(ctx context.Context, opts *ListListOptions) (*pagination.Paginated[List], error) {
	path := "api/shopping-list/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[List]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single shopping list by ID.
func (s *ListService) Get(ctx context.Context, id int) (*List, error) {
	var sl List
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/shopping-list/%d/", id), nil, &sl); err != nil {
		return nil, err
	}
	return &sl, nil
}

// Create creates a new shopping list.
func (s *ListService) Create(ctx context.Context, sl *List) (*List, error) {
	var result List
	if err := s.exec.DoJSON(ctx, "POST", "api/shopping-list/", sl, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a shopping list.
func (s *ListService) Update(ctx context.Context, sl *List) (*List, error) {
	var result List
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/shopping-list/%d/", sl.ID), sl, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a shopping list.
func (s *ListService) Patch(ctx context.Context, sl *List) (*List, error) {
	var result List
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/shopping-list/%d/", sl.ID), sl, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a shopping list.
func (s *ListService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/shopping-list/%d/", id), nil, nil)
}

// Cascading returns paginated objects that would be cascade-deleted.
func (s *ListService) Cascading(ctx context.Context, id int) (*pagination.Paginated[ModelReference], error) {
	var page pagination.Paginated[ModelReference]
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/shopping-list/%d/cascading/", id), nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns paginated objects whose fields would be set to NULL.
func (s *ListService) Nulling(ctx context.Context, id int) (*pagination.Paginated[ModelReference], error) {
	var page pagination.Paginated[ModelReference]
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/shopping-list/%d/nulling/", id), nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns paginated objects that would prevent deletion.
func (s *ListService) Protecting(ctx context.Context, id int) (*pagination.Paginated[ModelReference], error) {
	var page pagination.Paginated[ModelReference]
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/shopping-list/%d/protecting/", id), nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// EntryService provides access to Shopping List Entry API endpoints.
type EntryService struct {
	exec executor.Executor
}

// NewEntryService creates a new EntryService.
func NewEntryService(e executor.Executor) *EntryService {
	return &EntryService{exec: e}
}

// List returns a paginated list of shopping list entries.
func (s *EntryService) List(ctx context.Context, opts *ListEntryListOptions) (*pagination.Paginated[ListEntry], error) {
	path := "api/shopping-list-entry/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[ListEntry]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single shopping list entry by ID.
func (s *EntryService) Get(ctx context.Context, id int) (*ListEntry, error) {
	var entry ListEntry
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/shopping-list-entry/%d/", id), nil, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Create creates a new shopping list entry.
func (s *EntryService) Create(ctx context.Context, entry *ListEntry) (*ListEntry, error) {
	var result ListEntry
	if err := s.exec.DoJSON(ctx, "POST", "api/shopping-list-entry/", entry, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a shopping list entry.
func (s *EntryService) Update(ctx context.Context, entry *ListEntry) (*ListEntry, error) {
	var result ListEntry
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/shopping-list-entry/%d/", entry.ID), entry, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a shopping list entry.
func (s *EntryService) Patch(ctx context.Context, entry *ListEntry) (*ListEntry, error) {
	var result ListEntry
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/shopping-list-entry/%d/", entry.ID), entry, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a shopping list entry.
func (s *EntryService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/shopping-list-entry/%d/", id), nil, nil)
}

// BulkUpdate performs bulk operations on shopping list entries.
func (s *EntryService) BulkUpdate(ctx context.Context, bulk *ListEntryBulk) (*ListEntryBulk, error) {
	var result ListEntryBulk
	if err := s.exec.DoJSON(ctx, "POST", "api/shopping-list-entry/bulk/", bulk, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RecipeService provides access to Shopping List Recipe API endpoints.
type RecipeService struct {
	exec executor.Executor
}

// NewRecipeService creates a new RecipeService.
func NewRecipeService(e executor.Executor) *RecipeService {
	return &RecipeService{exec: e}
}

// List returns a paginated list of shopping list recipes.
func (s *RecipeService) List(ctx context.Context, opts *ListRecipeListOptions) (*pagination.Paginated[ListRecipe], error) {
	path := "api/shopping-list-recipe/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[ListRecipe]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single shopping list recipe by ID.
func (s *RecipeService) Get(ctx context.Context, id int) (*ListRecipe, error) {
	var slr ListRecipe
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/shopping-list-recipe/%d/", id), nil, &slr); err != nil {
		return nil, err
	}
	return &slr, nil
}

// Create creates a new shopping list recipe.
func (s *RecipeService) Create(ctx context.Context, slr *ListRecipe) (*ListRecipe, error) {
	var result ListRecipe
	if err := s.exec.DoJSON(ctx, "POST", "api/shopping-list-recipe/", slr, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a shopping list recipe.
func (s *RecipeService) Update(ctx context.Context, slr *ListRecipe) (*ListRecipe, error) {
	var result ListRecipe
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/shopping-list-recipe/%d/", slr.ID), slr, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a shopping list recipe.
func (s *RecipeService) Patch(ctx context.Context, slr *ListRecipe) (*ListRecipe, error) {
	var result ListRecipe
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/shopping-list-recipe/%d/", slr.ID), slr, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a shopping list recipe.
func (s *RecipeService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/shopping-list-recipe/%d/", id), nil, nil)
}

// BulkCreateEntries creates multiple entries from a shopping list recipe.
func (s *RecipeService) BulkCreateEntries(ctx context.Context, id int, req *ListEntryBulkCreate) (*ListEntryBulkCreate, error) {
	var result ListEntryBulkCreate
	if err := s.exec.DoJSON(ctx, "POST", fmt.Sprintf("api/shopping-list-recipe/%d/bulk_create_entries/", id), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
