package step

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Step API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new StepService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of steps.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Step], error) {
	path := "api/step/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Step]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a step by ID.
func (s *Service) Get(ctx context.Context, id int) (*Step, error) {
	var st Step
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/step/%d/", id), nil, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// Create creates a new step.
func (s *Service) Create(ctx context.Context, st *Step) (*Step, error) {
	var result Step
	if err := s.exec.DoJSON(ctx, "POST", "api/step/", st, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a step.
func (s *Service) Update(ctx context.Context, st *Step) (*Step, error) {
	var result Step
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/step/%d/", st.ID), st, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a step.
func (s *Service) Patch(ctx context.Context, st *Step) (*Step, error) {
	var result Step
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/step/%d/", st.ID), st, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a step by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/step/%d/", id), nil, nil)
}

// Cascading returns objects that would be cascade-deleted with this step.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Step], error) {
	path := fmt.Sprintf("api/step/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Step]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this step is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Step], error) {
	path := fmt.Sprintf("api/step/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Step]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this step from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Step], error) {
	path := fmt.Sprintf("api/step/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Step]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
