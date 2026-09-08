package mealplan

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Meal Plan API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new MealPlanService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of meal plans.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[MealPlan], error) {
	path := "api/meal-plan/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[MealPlan]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single meal plan by ID.
func (s *Service) Get(ctx context.Context, id int) (*MealPlan, error) {
	var mp MealPlan
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/meal-plan/%d/", id), nil, &mp); err != nil {
		return nil, err
	}
	return &mp, nil
}

// Create creates a new meal plan entry.
func (s *Service) Create(ctx context.Context, mp *MealPlan) (*MealPlan, error) {
	var result MealPlan
	if err := s.exec.DoJSON(ctx, "POST", "api/meal-plan/", mp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a meal plan entry.
func (s *Service) Update(ctx context.Context, mp *MealPlan) (*MealPlan, error) {
	var result MealPlan
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/meal-plan/%d/", mp.ID), mp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a meal plan entry.
func (s *Service) Patch(ctx context.Context, mp *MealPlan) (*MealPlan, error) {
	var result MealPlan
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/meal-plan/%d/", mp.ID), mp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a meal plan entry.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/meal-plan/%d/", id), nil, nil)
}

// ICAL returns an iCalendar feed for meal plans.
// The response body is a text/calendar string.
func (s *Service) ICAL(ctx context.Context, opts *pagination.ListOptions) (string, error) {
	path := "api/meal-plan/ical/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	resp, err := s.exec.DoRaw(ctx, "GET", path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API error %d", resp.StatusCode)
	}

	var buf strings.Builder
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MealTypeService provides access to Meal Type API endpoints.
type MealTypeService struct {
	exec executor.Executor
}

// NewMealTypeService creates a new MealTypeService.
func NewMealTypeService(e executor.Executor) *MealTypeService {
	return &MealTypeService{exec: e}
}

// List returns a paginated list of meal types.
func (s *MealTypeService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[MealType], error) {
	path := "api/meal-type/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[MealType]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single meal type by ID.
func (s *MealTypeService) Get(ctx context.Context, id int) (*MealType, error) {
	var mt MealType
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/meal-type/%d/", id), nil, &mt); err != nil {
		return nil, err
	}
	return &mt, nil
}

// Create creates a new meal type.
func (s *MealTypeService) Create(ctx context.Context, mt *MealType) (*MealType, error) {
	var result MealType
	if err := s.exec.DoJSON(ctx, "POST", "api/meal-type/", mt, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a meal type.
func (s *MealTypeService) Update(ctx context.Context, mt *MealType) (*MealType, error) {
	var result MealType
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/meal-type/%d/", mt.ID), mt, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a meal type.
func (s *MealTypeService) Patch(ctx context.Context, mt *MealType) (*MealType, error) {
	var result MealType
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/meal-type/%d/", mt.ID), mt, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a meal type.
func (s *MealTypeService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/meal-type/%d/", id), nil, nil)
}

// Cascading returns paginated objects that would be cascade-deleted.
func (s *MealTypeService) Cascading(ctx context.Context, id int) (*pagination.Paginated[ModelReference], error) {
	var page pagination.Paginated[ModelReference]
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/meal-type/%d/cascading/", id), nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Nulling returns paginated objects whose fields would be set to NULL.
func (s *MealTypeService) Nulling(ctx context.Context, id int) (*pagination.Paginated[ModelReference], error) {
	var page pagination.Paginated[ModelReference]
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/meal-type/%d/nulling/", id), nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Protecting returns paginated objects that would prevent deletion.
func (s *MealTypeService) Protecting(ctx context.Context, id int) (*pagination.Paginated[ModelReference], error) {
	var page pagination.Paginated[ModelReference]
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/meal-type/%d/protecting/", id), nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// AutoPlanService provides access to the Auto Plan API endpoint.
type AutoPlanService struct {
	exec executor.Executor
}

// NewAutoPlanService creates a new AutoPlanService.
func NewAutoPlanService(e executor.Executor) *AutoPlanService {
	return &AutoPlanService{exec: e}
}

// Plan creates a meal plan automatically based on keywords and date range.
func (s *AutoPlanService) Plan(ctx context.Context, req *AutoMealPlanRequest) (*AutoMealPlanRequest, error) {
	var result AutoMealPlanRequest
	if err := s.exec.DoJSON(ctx, "POST", "api/auto-meal-plan/", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ModelReference is a generic model reference used by cascading/nulling/protecting endpoints.
type ModelReference struct {
	ID    int    `json:"id"`
	Model string `json:"model"`
	Name  string `json:"name"`
}
