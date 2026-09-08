// Package tandoor provides a Go API client for Tandoor Recipes.
//
// Tandoor is a Django-based recipe management web application. This client
// communicates with the Tandoor REST API using OAuth2 access tokens.
//
// # Quick Start
//
//	client, err := tandoor.NewClient(
//		"https://recipes.example.com",
//		tandoor.WithAccessToken("tda_your_token_here"),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	recipes, err := client.Recipes().List(context.Background())
package tandoor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/swedishborgie/go-tandoor/action"
	"github.com/swedishborgie/go-tandoor/auth"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/importexport"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/inventory"
	"github.com/swedishborgie/go-tandoor/keyword"
	"github.com/swedishborgie/go-tandoor/mealplan"
	"github.com/swedishborgie/go-tandoor/misc"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/swedishborgie/go-tandoor/recipe"
	"github.com/swedishborgie/go-tandoor/recipebook"
	"github.com/swedishborgie/go-tandoor/shopping"
	"github.com/swedishborgie/go-tandoor/space"
	"github.com/swedishborgie/go-tandoor/step"
	"github.com/swedishborgie/go-tandoor/storage"
	"github.com/swedishborgie/go-tandoor/supermarket"
	"github.com/swedishborgie/go-tandoor/unit"
)

// Paginated is a type alias for pagination.Paginated[T], re-exported here
// so that service methods return types from the root package directly.
type Paginated[T any] = pagination.Paginated[T]

// Default values.
const (
	DefaultTimeout = 30 * time.Second
)

// Client is the Tandoor API client.
type Client struct {
	httpClient  *http.Client
	baseURL     *url.URL
	accessToken string
	userAgent   string
}

// ClientOption configures a Client via the functional options pattern.
type ClientOption func(*Client) error

// defaultClient returns a Client with sensible defaults.
func defaultClient(baseURL string) *Client {
	u, _ := url.Parse(baseURL)
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:   u,
		userAgent: "go-tandoor/dev (+https://github.com/swedishborgie/go-tandoor)",
	}
}

// NewClient creates a new Tandoor API client.
//
// baseURL should include the scheme and host, e.g. "https://recipes.example.com".
// A trailing slash is optional and will be normalized.
func NewClient(baseURL string, opts ...ClientOption) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("tandoor: baseURL is required")
	}

	c := defaultClient(baseURL)
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// WithAccessToken sets the OAuth2 access token used for authentication.
//
// Tokens are obtained via POST /api-token-auth/ or POST /api/access-token/.
// The token is sent as "Authorization: Bearer <token>" on every request.
func WithAccessToken(token string) ClientOption {
	return func(c *Client) error {
		if token == "" {
			return fmt.Errorf("tandoor: access token is required")
		}
		c.accessToken = token
		return nil
	}
}

// WithHTTPClient sets a custom http.Client for all HTTP requests.
//
// This allows callers to configure timeouts, TLS settings, transport
// options, or middleware (e.g., retry logic).
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("tandoor: http.Client must not be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithBaseURL overrides the base URL.
//
// Typically used when the caller needs to change the base URL after
// initial construction (e.g., for multi-tenancy).
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		if baseURL == "" {
			return fmt.Errorf("tandoor: base URL is required")
		}
		u, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("tandoor: invalid base URL %q: %w", baseURL, err)
		}
		c.baseURL = u
		return nil
	}
}

// WithUserAgent sets a custom User-Agent header for all requests.
// If not set, a sensible default is used.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) error {
		if userAgent == "" {
			return fmt.Errorf("tandoor: user agent must not be empty")
		}
		c.userAgent = userAgent
		return nil
	}
}

// newRequest creates an *http.Request for the given method and path.
//
// The path is joined with the client's base URL. If body is non-nil it
// is JSON-encoded and the Content-Type header is set.
// The Authorization header is set if an access token is configured.
// newRequest returns the newRequest service.
func (c *Client) newRequest(ctx context.Context, method string, subPath string, body any) (*http.Request, error) {
	// Split query string from the path if present.
	subPath, query := splitQuery(subPath)

	// Build the path, normalizing separators and ensuring a trailing slash
	// (DRF requires trailing slashes on all endpoints).
	p := "/"
	if subPath != "" {
		p = path.Join(p, subPath)
	}
	// path.Join strips trailing slashes, so add one back.
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}

	u := c.baseURL.ResolveReference(&url.URL{
		Path:     p,
		RawQuery: query,
	})

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("tandoor: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("tandoor: create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

// newRequestNoAuth is like newRequest but omits the Authorization header.
// Used for auth endpoints that must not carry a Bearer token.
// newRequestNoAuth returns the newRequestNoAuth service.
func (c *Client) newRequestNoAuth(ctx context.Context, method string, subPath string, body any) (*http.Request, error) {
	subPath, query := splitQuery(subPath)

	p := "/"
	if subPath != "" {
		p = path.Join(p, subPath)
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}

	u := c.baseURL.ResolveReference(&url.URL{
		Path:     p,
		RawQuery: query,
	})

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("tandoor: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("tandoor: create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Note: no Authorization header set
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

// do executes an HTTP request and decodes the JSON response into v.
//
// If the response status is not in the 2xx range, a *TandoorError is
// returned. If v is nil the response body is discarded.
// do returns the do service.
func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tandoor: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("tandoor: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errObj := &TandoorError{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
		// Attempt to parse DRF error response.
		var raw map[string]any
		if json.Unmarshal(body, &raw) == nil {
			apiErr := &APIError{
				FieldErrors: make(map[string][]string),
			}
			if d, ok := raw["detail"]; ok {
				if ds, ok := d.(string); ok {
					apiErr.Detail = ds
				}
			}
			if nfe, ok := raw["non_field_errors"]; ok {
				if arr, ok := nfe.([]any); ok {
					for _, v := range arr {
						if s, ok := v.(string); ok {
							apiErr.NonFieldErrors = append(apiErr.NonFieldErrors, s)
						}
					}
				}
			}
			for k, v := range raw {
				if k == "detail" || k == "non_field_errors" {
					continue
				}
				if arr, ok := v.([]any); ok {
					strs := make([]string, 0, len(arr))
					for _, item := range arr {
						if s, ok := item.(string); ok {
							strs = append(strs, s)
						}
					}
					apiErr.FieldErrors[k] = strs
				} else if s, ok := v.(string); ok {
					apiErr.FieldErrors[k] = []string{s}
				}
			}
			switch {
			case apiErr.Detail != "":
				errObj.Message = apiErr.Detail
			case len(apiErr.NonFieldErrors) > 0:
				errObj.Message = strings.Join(apiErr.NonFieldErrors, "; ")
			case len(apiErr.FieldErrors) > 0:
				for _, errs := range apiErr.FieldErrors {
					if len(errs) > 0 {
						errObj.Message = errs[0]
						break
					}
				}
			}
			if apiErr.Detail != "" || len(apiErr.NonFieldErrors) > 0 || len(apiErr.FieldErrors) > 0 {
				errObj.apiErr = apiErr
			}
		}
		return errObj
	}

	if v != nil && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("tandoor: unmarshal response: %w", err)
		}
	}

	return nil
}

// doJSON is a convenience wrapper that creates a request and executes it.
// doJSON returns the doJSON service.
func (c *Client) doJSON(ctx context.Context, method string, subPath string, body any, v any) error {
	req, err := c.newRequest(ctx, method, subPath, body)
	if err != nil {
		return err
	}
	return c.do(req, v)
}

// splitQuery splits a path that may contain a query string ("?...") into
// the path prefix and the raw query string.
func splitQuery(raw string) (path, query string) {
	idx := strings.Index(raw, "?")
	if idx == -1 {
		return raw, ""
	}
	return raw[:idx], raw[idx+1:]
}

// DoJSON, DoRaw, and BaseURLOrigin implement the executor.Executor interface.
// These methods are exported so the Client satisfies the internal Executor interface
// that service packages depend on.

// DoJSON creates a request, executes it, and decodes the JSON response into v.
// DoJSON returns the DoJSON service.
func (c *Client) DoJSON(ctx context.Context, method, path string, body any, v any) error {
	return c.doJSON(ctx, method, path, body, v)
}

// DoRaw creates a request and executes it, returning the raw *http.Response.
// The caller is responsible for reading and closing the response body.
// DoRaw returns the DoRaw service.
func (c *Client) DoRaw(ctx context.Context, method, path string, body any) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

// DoRawNoAuth creates a request WITHOUT the Authorization header and
// executes it, returning the raw *http.Response.
// DoRawNoAuth returns the DoRawNoAuth service.
func (c *Client) DoRawNoAuth(ctx context.Context, method, path string, body any) (*http.Response, error) {
	req, err := c.newRequestNoAuth(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

// BaseURLOrigin returns the origin URL (scheme + host).
// BaseURLOrigin returns the BaseURLOrigin service.
func (c *Client) BaseURLOrigin() string {
	return c.baseURL.Scheme + "://" + c.baseURL.Host
}

// --- Service Accessors ---
// Each accessor returns a service configured with this client as executor.

// Auth returns the Auth service.
func (c *Client) Auth() *auth.Service {
	return auth.NewService(c)
}

// Foods returns the Foods service.
func (c *Client) Foods() *food.Service {
	return food.NewService(c)
}

// Ingredients returns the Ingredients service.
func (c *Client) Ingredients() *ingredient.Service {
	return ingredient.NewService(c)
}

// Keywords returns the Keywords service.
func (c *Client) Keywords() *keyword.Service {
	return keyword.NewService(c)
}

// Steps returns the Steps service.
func (c *Client) Steps() *step.Service {
	return step.NewService(c)
}

// Units returns the Units service.
func (c *Client) Units() *unit.Service {
	return unit.NewService(c)
}

// UnitConversions returns the UnitConversions service.
func (c *Client) UnitConversions() *unit.ConversionService {
	return unit.NewConversionService(c)
}

// Recipes returns the Recipes service.
func (c *Client) Recipes() *recipe.Service {
	return recipe.NewService(c)
}

// RecipeBooks returns the RecipeBooks service.
func (c *Client) RecipeBooks() *recipebook.Service {
	return recipebook.NewService(c)
}

// Properties returns the Properties service.
func (c *Client) Properties() *property.Service {
	return property.NewService(c)
}

// Inventory returns the Inventory service.
func (c *Client) Inventory() *inventory.Service {
	return inventory.NewService(c)
}

// InventoryLocations returns the InventoryLocations service.
func (c *Client) InventoryLocations() *inventory.LocationService {
	return inventory.NewLocationService(c)
}

// Storages returns the Storages service.
func (c *Client) Storages() *storage.Service {
	return storage.NewService(c)
}

// Spaces returns the Spaces service.
func (c *Client) Spaces() *space.Service {
	return space.NewService(c)
}

// Users returns the Users service.
func (c *Client) Users() *space.UserService {
	return space.NewUserService(c)
}

// Groups returns the Groups service.
func (c *Client) Groups() *space.GroupService {
	return space.NewGroupService(c)
}

// Households returns the Households service.
func (c *Client) Households() *space.HouseholdService {
	return space.NewHouseholdService(c)
}

// UserSpaces returns the UserSpaces service.
func (c *Client) UserSpaces() *space.UserSpaceService {
	return space.NewUserSpaceService(c)
}

// UserPreferences returns the UserPreferences service.
func (c *Client) UserPreferences() *space.UserPreferenceService {
	return space.NewUserPreferenceService(c)
}

// InviteLinks returns the InviteLinks service.
func (c *Client) InviteLinks() *space.InviteLinkService {
	return space.NewInviteLinkService(c)
}

// MealPlans returns the MealPlans service.
func (c *Client) MealPlans() *mealplan.Service {
	return mealplan.NewService(c)
}

// MealTypes returns the MealTypes service.
func (c *Client) MealTypes() *mealplan.MealTypeService {
	return mealplan.NewMealTypeService(c)
}

// AutoPlan returns the AutoPlan service.
func (c *Client) AutoPlan() *mealplan.AutoPlanService {
	return mealplan.NewAutoPlanService(c)
}

// ShoppingLists returns the ShoppingLists service.
func (c *Client) ShoppingLists() *shopping.ListService {
	return shopping.NewListService(c)
}

// ShoppingEntries returns the ShoppingEntries service.
func (c *Client) ShoppingEntries() *shopping.EntryService {
	return shopping.NewEntryService(c)
}

// ShoppingRecipes returns the ShoppingRecipes service.
func (c *Client) ShoppingRecipes() *shopping.RecipeService {
	return shopping.NewRecipeService(c)
}

// Supermarkets returns the Supermarkets service.
func (c *Client) Supermarkets() *supermarket.Service {
	return supermarket.NewService(c)
}

// SupermarketCategories returns the SupermarketCategories service.
func (c *Client) SupermarketCategories() *supermarket.CategoryService {
	return supermarket.NewCategoryService(c)
}

// SupermarketCategoryRelations returns the SupermarketCategoryRelations service.
func (c *Client) SupermarketCategoryRelations() *supermarket.CategoryRelationService {
	return supermarket.NewCategoryRelationService(c)
}

// RecipeFromSource returns the RecipeFromSource service.
func (c *Client) RecipeFromSource() *action.RecipeFromSourceService {
	return action.NewRecipeFromSourceService(c)
}

// AiImport returns the AiImport service.
func (c *Client) AiImport() *action.AiImportService {
	return action.NewAiImportService(c)
}

// AiStepSort returns the AiStepSort service.
func (c *Client) AiStepSort() *action.AiStepSortService {
	return action.NewAiStepSortService(c)
}

// IngredientParser returns the IngredientParser service.
func (c *Client) IngredientParser() *action.IngredientParserService {
	return action.NewIngredientParserService(c)
}

// FdcSearch returns the FdcSearch service.
func (c *Client) FdcSearch() *action.FdcSearchService {
	return action.NewFdcSearchService(c)
}

// ImportOpenData returns the ImportOpenData service.
func (c *Client) ImportOpenData() *action.ImportOpenDataService {
	return action.NewImportOpenDataService(c)
}

// ShareLinks returns the ShareLinks service.
func (c *Client) ShareLinks() *action.ShareLinkService {
	return action.NewShareLinkService(c)
}

// RecipeImports returns the RecipeImports service.
func (c *Client) RecipeImports() *importexport.RecipeImportService {
	return importexport.NewRecipeImportService(c)
}

// ImportLogs returns the ImportLogs service.
func (c *Client) ImportLogs() *importexport.ImportLogService {
	return importexport.NewImportLogService(c)
}

// ExportLogs returns the ExportLogs service.
func (c *Client) ExportLogs() *importexport.ExportLogService {
	return importexport.NewExportLogService(c)
}

// BookmarkletImports returns the BookmarkletImports service.
func (c *Client) BookmarkletImports() *importexport.BookmarkletImportService {
	return importexport.NewBookmarkletImportService(c)
}

// Syncs returns the Syncs service.
func (c *Client) Syncs() *importexport.SyncService {
	return importexport.NewSyncService(c)
}

// SyncLogs returns the SyncLogs service.
func (c *Client) SyncLogs() *importexport.SyncLogService {
	return importexport.NewSyncLogService(c)
}

// AppExport returns the AppExport service.
func (c *Client) AppExport() *importexport.AppExportService {
	return importexport.NewAppExportService(c)
}

// CookLogs returns the CookLogs service.
func (c *Client) CookLogs() *misc.Service {
	return misc.NewService(c)
}

// ViewLogs returns the ViewLogs service.
func (c *Client) ViewLogs() *misc.ViewLogService {
	return misc.NewViewLogService(c)
}

// UserFiles returns the UserFiles service.
func (c *Client) UserFiles() *misc.UserFileService {
	return misc.NewUserFileService(c)
}

// Automations returns the Automations service.
func (c *Client) Automations() *misc.AutomationService {
	return misc.NewAutomationService(c)
}

// CustomFilters returns the CustomFilters service.
func (c *Client) CustomFilters() *misc.CustomFilterService {
	return misc.NewCustomFilterService(c)
}

// ConnectorConfigs returns the ConnectorConfigs service.
func (c *Client) ConnectorConfigs() *misc.ConnectorConfigService {
	return misc.NewConnectorConfigService(c)
}

// SearchFields returns the SearchFields service.
func (c *Client) SearchFields() *misc.SearchFieldsService {
	return misc.NewSearchFieldsService(c)
}

// SearchPreference returns the SearchPreference service.
func (c *Client) SearchPreference() *misc.SearchPreferenceService {
	return misc.NewSearchPreferenceService(c)
}

// AiProviders returns the AiProviders service.
func (c *Client) AiProviders() *misc.AiProviderService {
	return misc.NewAiProviderService(c)
}

// AiLogs returns the AiLogs service.
func (c *Client) AiLogs() *misc.AiLogService {
	return misc.NewAiLogService(c)
}

// Localization returns the Localization service.
func (c *Client) Localization() *misc.LocalizationService {
	return misc.NewLocalizationService(c)
}

// ServerSettings returns the ServerSettings service.
func (c *Client) ServerSettings() *misc.ServerSettingsService {
	return misc.NewServerSettingsService(c)
}
