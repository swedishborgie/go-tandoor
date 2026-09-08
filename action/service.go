package action

import (
	"context"
	"fmt"
	"net/url"

	"github.com/swedishborgie/go-tandoor/internal/executor"
)

// RecipeFromSourceService provides access to the recipe-from-source endpoint.
type RecipeFromSourceService struct {
	exec executor.Executor
}

// NewRecipeFromSourceService creates a new RecipeFromSourceService.
func NewRecipeFromSourceService(e executor.Executor) *RecipeFromSourceService {
	return &RecipeFromSourceService{exec: e}
}

// Import imports a recipe from a URL, HTML data, or bookmarklet.
func (s *RecipeFromSourceService) Import(ctx context.Context, req *RecipeFromSourceRequest) (*RecipeFromSourceResponse, error) {
	var resp RecipeFromSourceResponse
	if err := s.exec.DoJSON(ctx, "POST", "api/recipe-from-source/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AiImportService provides access to the AI import endpoint.
type AiImportService struct {
	exec executor.Executor
}

// NewAiImportService creates a new AiImportService.
func NewAiImportService(e executor.Executor) *AiImportService {
	return &AiImportService{exec: e}
}

// Import converts an image, PDF, or text to a structured recipe using AI.
func (s *AiImportService) Import(ctx context.Context, req *AiImportRequest) (*RecipeFromSourceResponse, error) {
	var resp RecipeFromSourceResponse
	if err := s.exec.DoJSON(ctx, "POST", "api/ai-import/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AiStepSortService provides access to the AI step sort endpoint.
type AiStepSortService struct {
	exec executor.Executor
}

// NewAiStepSortService creates a new AiStepSortService.
func NewAiStepSortService(e executor.Executor) *AiStepSortService {
	return &AiStepSortService{exec: e}
}

// Sort uses AI to sort recipe ingredients to the appropriate steps.
func (s *AiStepSortService) Sort(ctx context.Context, recipe map[string]any, providerID int) (map[string]any, error) {
	path := fmt.Sprintf("api/ai-step-sort/?provider=%d", providerID)
	var result map[string]any
	if err := s.exec.DoJSON(ctx, "POST", path, recipe, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// IngredientParserService provides access to the ingredient parser endpoint.
type IngredientParserService struct {
	exec executor.Executor
}

// NewIngredientParserService creates a new IngredientParserService.
func NewIngredientParserService(e executor.Executor) *IngredientParserService {
	return &IngredientParserService{exec: e}
}

// Parse parses one or more ingredient strings into structured data.
func (s *IngredientParserService) Parse(ctx context.Context, req *IngredientParserRequest) (*IngredientParserResponse, error) {
	var resp IngredientParserResponse
	if err := s.exec.DoJSON(ctx, "POST", "api/ingredient-parser/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FdcSearchService provides access to the USDA FoodData Central search endpoint.
type FdcSearchService struct {
	exec executor.Executor
}

// NewFdcSearchService creates a new FdcSearchService.
func NewFdcSearchService(e executor.Executor) *FdcSearchService {
	return &FdcSearchService{exec: e}
}

// Search searches the USDA FoodData Central API.
func (s *FdcSearchService) Search(ctx context.Context, query string, dataTypes []string) (*FdcQueryResponse, error) {
	v := url.Values{}
	v.Set("query", query)
	for _, dt := range dataTypes {
		v.Add("dataType", dt)
	}
	path := fmt.Sprintf("api/fdc-search/?%s", v.Encode())
	var resp FdcQueryResponse
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ImportOpenDataService provides access to the open data import endpoint.
type ImportOpenDataService struct {
	exec executor.Executor
}

// NewImportOpenDataService creates a new ImportOpenDataService.
func NewImportOpenDataService(e executor.Executor) *ImportOpenDataService {
	return &ImportOpenDataService{exec: e}
}

// GetMetadata fetches available open data versions and datatypes.
func (s *ImportOpenDataService) GetMetadata(ctx context.Context) (*ImportOpenDataMetadata, error) {
	var resp ImportOpenDataMetadata
	if err := s.exec.DoJSON(ctx, "GET", "api/import-open-data/", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Import imports open data for the specified version and datatypes.
func (s *ImportOpenDataService) Import(ctx context.Context, req *ImportOpenDataRequest) (*ImportOpenDataResponse, error) {
	var resp ImportOpenDataResponse
	if err := s.exec.DoJSON(ctx, "POST", "api/import-open-data/", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ShareLinkService provides access to the share link endpoint.
type ShareLinkService struct {
	exec executor.Executor
}

// NewShareLinkService creates a new ShareLinkService.
func NewShareLinkService(e executor.Executor) *ShareLinkService {
	return &ShareLinkService{exec: e}
}

// Create creates a shareable link for a recipe.
func (s *ShareLinkService) Create(ctx context.Context, recipeID int) (*ShareLinkResponse, error) {
	var resp ShareLinkResponse
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/share-link/%d", recipeID), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
