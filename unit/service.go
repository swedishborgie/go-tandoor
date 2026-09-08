package unit

import (
	"context"
	"fmt"
	"net/url"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Unit API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new UnitService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of units.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Unit], error) {
	path := "api/unit/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Unit]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a unit by ID.
func (s *Service) Get(ctx context.Context, id int) (*Unit, error) {
	var u Unit
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/unit/%d/", id), nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Create creates a new unit.
func (s *Service) Create(ctx context.Context, u *Unit) (*Unit, error) {
	var result Unit
	if err := s.exec.DoJSON(ctx, "POST", "api/unit/", u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a unit.
func (s *Service) Update(ctx context.Context, u *Unit) (*Unit, error) {
	var result Unit
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/unit/%d/", u.ID), u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a unit.
func (s *Service) Patch(ctx context.Context, u *Unit) (*Unit, error) {
	var result Unit
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/unit/%d/", u.ID), u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a unit by ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/unit/%d/", id), nil, nil)
}

// Merge merges sourceID into targetID and returns the merged unit.
func (s *Service) Merge(ctx context.Context, sourceID, targetID int) (*Unit, error) {
	var result Unit
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/unit/%d/merge/%d/", sourceID, targetID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Cascading returns objects that would be cascade-deleted with this unit.
func (s *Service) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Unit], error) {
	path := fmt.Sprintf("api/unit/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Unit]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this unit is deleted.
func (s *Service) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Unit], error) {
	path := fmt.Sprintf("api/unit/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Unit]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this unit from deletion.
func (s *Service) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Unit], error) {
	path := fmt.Sprintf("api/unit/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Unit]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// ConversionService provides access to Unit Conversion API endpoints.
type ConversionService struct {
	exec executor.Executor
}

// NewConversionService creates a new ConversionService.
func NewConversionService(e executor.Executor) *ConversionService {
	return &ConversionService{exec: e}
}

// List returns a paginated list of unit conversions.
func (s *ConversionService) List(ctx context.Context, opts *ConversionListOptions) (*pagination.Paginated[Conversion], error) {
	path := "api/unit-conversion/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Conversion]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a unit conversion by ID.
func (s *ConversionService) Get(ctx context.Context, id int) (*Conversion, error) {
	var conv Conversion
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/unit-conversion/%d/", id), nil, &conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

func convToPayload(conv *Conversion) map[string]any {
	p := map[string]any{
		"base_amount":      conv.BaseAmount,
		"converted_amount": conv.ConvertedAmount,
	}
	if conv.BaseUnit != nil {
		p["base_unit"] = conv.BaseUnit.ID
	}
	if conv.ConvertedUnit != nil {
		p["converted_unit"] = conv.ConvertedUnit.ID
	}
	if conv.Food != nil {
		p["food"] = conv.Food.ID
	}
	if conv.OpenDataSlug != "" {
		p["open_data_slug"] = conv.OpenDataSlug
	}
	return p
}

// Create creates a new unit conversion.
func (s *ConversionService) Create(ctx context.Context, conv *Conversion) (*Conversion, error) {
	var result Conversion
	if err := s.exec.DoJSON(ctx, "POST", "api/unit-conversion/", convToPayload(conv), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of a unit conversion.
func (s *ConversionService) Update(ctx context.Context, conv *Conversion) (*Conversion, error) {
	var result Conversion
	payload := convToPayload(conv)
	if conv.ID != 0 {
		payload["id"] = conv.ID
	}
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/unit-conversion/%d/", conv.ID), payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of a unit conversion.
func (s *ConversionService) Patch(ctx context.Context, conv *Conversion) (*Conversion, error) {
	var result Conversion
	payload := convToPayload(conv)
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/unit-conversion/%d/", conv.ID), payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a unit conversion by ID.
func (s *ConversionService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/unit-conversion/%d/", id), nil, nil)
}

// ConversionListOptions holds optional query parameters for listing unit conversions.
type ConversionListOptions struct {
	*pagination.ListOptions
	FoodID *int `url:"food_id"`
}

// QueryString returns the encoded query string for this options struct.
func (o *ConversionListOptions) QueryString() string {
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
