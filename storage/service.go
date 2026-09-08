package storage

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Storage API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new StorageService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of storages.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Storage], error) {
	path := "api/storage/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Storage]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a storage by ID.
func (s *Service) Get(ctx context.Context, id int) (*Storage, error) {
	var st Storage
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/storage/%d/", id), nil, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// Create creates a new storage.
func (s *Service) Create(ctx context.Context, st *Storage) (*Storage, error) {
	var result Storage
	if err := s.exec.DoJSON(ctx, "POST", "api/storage/", st, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a storage.
func (s *Service) Update(ctx context.Context, st *Storage) (*Storage, error) {
	var result Storage
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/storage/%d/", st.ID), st, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a storage.
func (s *Service) Patch(ctx context.Context, st *Storage) (*Storage, error) {
	var result Storage
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/storage/%d/", st.ID), st, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a storage by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/storage/%d/", id), nil, nil)
}

// Cascading returns objects that would be cascade-deleted with this storage.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Storage], error) {
	path := fmt.Sprintf("api/storage/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Storage]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this storage is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Storage], error) {
	path := fmt.Sprintf("api/storage/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Storage]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this storage from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Storage], error) {
	path := fmt.Sprintf("api/storage/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Storage]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
