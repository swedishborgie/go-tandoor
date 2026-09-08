package inventory

import (
	"context"
	"fmt"
	"net/url"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Inventory API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new InventoryService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// LocationService provides access to Inventory Location API endpoints.
type LocationService struct {
	exec executor.Executor
}

// NewLocationService creates a new LocationService.
func NewLocationService(e executor.Executor) *LocationService {
	return &LocationService{exec: e}
}

// List returns a paginated list of inventory locations.
func (s *LocationService) List(ctx context.Context, opts *LocationListOptions) (*pagination.Paginated[Location], error) {
	path := "api/inventory-location/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Location]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves an inventory location by ID.
func (s *LocationService) Get(ctx context.Context, id int) (*Location, error) {
	var loc Location
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/inventory-location/%d/", id), nil, &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

// Create creates a new inventory location.
func (s *LocationService) Create(ctx context.Context, loc *Location) (*Location, error) {
	var result Location
	if err := s.exec.DoJSON(ctx, "POST", "api/inventory-location/", loc, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of an inventory location.
func (s *LocationService) Update(ctx context.Context, loc *Location) (*Location, error) {
	var result Location
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/inventory-location/%d/", loc.ID), loc, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of an inventory location.
func (s *LocationService) Patch(ctx context.Context, loc *Location) (*Location, error) {
	var result Location
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/inventory-location/%d/", loc.ID), loc, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes an inventory location by ID.
func (s *LocationService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/inventory-location/%d/", id), nil, nil)
}

// Cascading returns objects that would be cascade-deleted with this inventory location.
func (s *LocationService) Cascading(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Location], error) {
	path := fmt.Sprintf("api/inventory-location/%d/cascading/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Location]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns objects that would be nullified if this inventory location is deleted.
func (s *LocationService) Nulling(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Location], error) {
	path := fmt.Sprintf("api/inventory-location/%d/nulling/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Location]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns objects that would protect this inventory location from deletion.
func (s *LocationService) Protecting(ctx context.Context, id int, opts *pagination.ListOptions) (*pagination.Paginated[Location], error) {
	path := fmt.Sprintf("api/inventory-location/%d/protecting/", id)
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Location]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// EntryService provides access to Inventory Entry API endpoints.
type EntryService struct {
	exec executor.Executor
}

// NewEntryService creates a new EntryService.
func NewEntryService(e executor.Executor) *EntryService {
	return &EntryService{exec: e}
}

// List returns a paginated list of inventory entries.
func (s *EntryService) List(ctx context.Context, opts *EntryListOptions) (*pagination.Paginated[Entry], error) {
	path := "api/inventory-entry/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Entry]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves an inventory entry by ID.
func (s *EntryService) Get(ctx context.Context, id int) (*Entry, error) {
	var entry Entry
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/inventory-entry/%d/", id), nil, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Create creates a new inventory entry.
func (s *EntryService) Create(ctx context.Context, entry *Entry) (*Entry, error) {
	var result Entry
	if err := s.exec.DoJSON(ctx, "POST", "api/inventory-entry/", entry, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update performs a full update of an inventory entry.
func (s *EntryService) Update(ctx context.Context, entry *Entry) (*Entry, error) {
	var result Entry
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/inventory-entry/%d/", entry.ID), entry, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update of an inventory entry.
func (s *EntryService) Patch(ctx context.Context, entry *Entry) (*Entry, error) {
	var result Entry
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/inventory-entry/%d/", entry.ID), entry, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes an inventory entry by ID.
func (s *EntryService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/inventory-entry/%d/", id), nil, nil)
}

// EntryListOptions holds optional query parameters for listing inventory entries.
type EntryListOptions struct {
	*pagination.ListOptions
	IncludeEmpty *bool   `url:"empty"`
	Code         *string `url:"code"`
	FoodID       *int    `url:"food_id"`
	LocationID   *int    `url:"inventory_location_id"`
}

// Empty returns true if no options are set.
func (o *EntryListOptions) Empty() bool {
	return o == nil || (o.ListOptions == nil && o.IncludeEmpty == nil && o.Code == nil && o.FoodID == nil && o.LocationID == nil)
}

// QueryString returns the encoded query string.
func (o *EntryListOptions) QueryString() string {
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
	if o.IncludeEmpty != nil {
		v.Set("empty", fmt.Sprintf("%v", *o.IncludeEmpty))
	}
	if o.Code != nil {
		v.Set("code", *o.Code)
	}
	if o.FoodID != nil {
		v.Set("food_id", fmt.Sprintf("%d", *o.FoodID))
	}
	if o.LocationID != nil {
		v.Set("inventory_location_id", fmt.Sprintf("%d", *o.LocationID))
	}
	return v.Encode()
}

// LogService provides access to Inventory Log API endpoints.
type LogService struct {
	exec executor.Executor
}

// NewLogService creates a new LogService.
func NewLogService(e executor.Executor) *LogService {
	return &LogService{exec: e}
}

// List returns a paginated list of inventory logs.
func (s *LogService) List(ctx context.Context, opts *LogListOptions) (*pagination.Paginated[Log], error) {
	path := "api/inventory-log/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Log]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves an inventory log by ID.
func (s *LogService) Get(ctx context.Context, id int) (*Log, error) {
	var log Log
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/inventory-log/%d/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// LogListOptions holds optional query parameters for listing inventory logs.
type LogListOptions struct {
	*pagination.ListOptions
	EntryID *int `url:"entry_id"`
	FoodID  *int `url:"food_id"`
}

// QueryString returns the encoded query string.
func (o *LogListOptions) QueryString() string {
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
	if o.EntryID != nil {
		v.Set("entry_id", fmt.Sprintf("%d", *o.EntryID))
	}
	if o.FoodID != nil {
		v.Set("food_id", fmt.Sprintf("%d", *o.FoodID))
	}
	return v.Encode()
}
