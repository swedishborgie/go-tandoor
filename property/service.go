package property

import (
	"context"
	"fmt"
	"net/url"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Property API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new PropertyService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of properties.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Property], error) {
	path := "api/property/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Property]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a property by ID.
func (s *Service) Get(ctx context.Context, id int) (*Property, error) {
	var p Property
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/property/%d/", id), nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Create creates a new property.
func (s *Service) Create(ctx context.Context, p *Property) (*Property, error) {
	var result Property
	if err := s.exec.DoJSON(ctx, "POST", "api/property/", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a property.
func (s *Service) Update(ctx context.Context, p *Property) (*Property, error) {
	var result Property
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/property/%d/", p.ID), p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a property.
func (s *Service) Patch(ctx context.Context, p *Property) (*Property, error) {
	var result Property
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/property/%d/", p.ID), p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a property by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/property/%d/", id), nil, nil)
}

// Cascading returns objects that would be cascade-deleted with this property.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Property], error) {
	path := fmt.Sprintf("api/property/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Property]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this property is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Property], error) {
	path := fmt.Sprintf("api/property/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Property]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this property from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Property], error) {
	path := fmt.Sprintf("api/property/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Property]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// ListOptions holds optional query parameters for listing properties.
type ListOptions struct {
	*pagination.ListOptions
	FoodID *int `url:"food_id"`
}

// QueryString returns the encoded query string for this options struct.
func (o *ListOptions) QueryString() string {
	if o == nil {
		return ""
	}
	v := url.Values{}
	if o.ListOptions != nil && o.ListOptions.QueryString() != "" {
		qv, _ := url.ParseQuery(o.ListOptions.QueryString())
		for k, vals := range qv {
			for _, val := range vals {
				v.Add(k, val)
			}
		}
	}
	if o.FoodID != nil {
		v.Set("food_id", fmt.Sprintf("%d", *o.FoodID))
	}
	return v.Encode()
}

// TypeService provides access to Property Type API endpoints.
type TypeService struct {
	exec executor.Executor
}

// NewTypeService creates a new TypeService.
func NewTypeService(e executor.Executor) *TypeService {
	return &TypeService{exec: e}
}

// List returns a paginated list of property types.
func (s *TypeService) List(ctx context.Context, opts *TypeListOptions) (*pagination.Paginated[Type], error) {
	path := "api/property-type/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Type]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a property type by ID.
func (s *TypeService) Get(ctx context.Context, id int) (*Type, error) {
	var p Type
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/property-type/%d/", id), nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Create creates a new property type.
func (s *TypeService) Create(ctx context.Context, p *Type) (*Type, error) {
	var result Type
	if err := s.exec.DoJSON(ctx, "POST", "api/property-type/", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a property type.
func (s *TypeService) Update(ctx context.Context, p *Type) (*Type, error) {
	var result Type
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/property-type/%d/", p.ID), p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a property type.
func (s *TypeService) Patch(ctx context.Context, p *Type) (*Type, error) {
	var result Type
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/property-type/%d/", p.ID), p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a property type by ID.
func (s *TypeService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/property-type/%d/", id), nil, nil)
}

// Cascading returns objects that would be cascade-deleted with this property type.
func (s *TypeService) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Type], error) {
	path := fmt.Sprintf("api/property-type/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Type]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this property type is deleted.
func (s *TypeService) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Type], error) {
	path := fmt.Sprintf("api/property-type/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Type]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this property type from deletion.
func (s *TypeService) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Type], error) {
	path := fmt.Sprintf("api/property-type/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Type]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
