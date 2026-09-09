package misc

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Cook Log API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new CookLogService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of cook logs.
func (s *Service) List(ctx context.Context, opts *CookLogListOptions) (*pagination.Paginated[CookLog], error) {
	path := "api/cook-log/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[CookLog]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single cook log by ID.
func (s *Service) Get(ctx context.Context, id int) (*CookLog, error) {
	var log CookLog
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/cook-log/%d/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// Create creates a new cook log entry.
func (s *Service) Create(ctx context.Context, log *CookLog) (*CookLog, error) {
	var result CookLog
	if err := s.exec.DoJSON(ctx, "POST", "api/cook-log/", log, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a cook log entry.
func (s *Service) Update(ctx context.Context, log *CookLog) (*CookLog, error) {
	var result CookLog
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/cook-log/%d/", log.ID), log, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a cook log entry.
func (s *Service) Patch(ctx context.Context, log *CookLog) (*CookLog, error) {
	var result CookLog
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/cook-log/%d/", log.ID), log, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a cook log entry.
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/cook-log/%d/", id), nil, nil)
}

// ViewLogService provides access to View Log API endpoints.
type ViewLogService struct {
	exec executor.Executor
}

// NewViewLogService creates a new ViewLogService.
func NewViewLogService(e executor.Executor) *ViewLogService {
	return &ViewLogService{exec: e}
}

// List returns a paginated list of view logs.
func (s *ViewLogService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[ViewLog], error) {
	path := "api/view-log/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[ViewLog]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single view log by ID.
func (s *ViewLogService) Get(ctx context.Context, id int) (*ViewLog, error) {
	var log ViewLog
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/view-log/%d/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// Create creates a new view log entry.
func (s *ViewLogService) Create(ctx context.Context, log *ViewLog) (*ViewLog, error) {
	var result ViewLog
	if err := s.exec.DoJSON(ctx, "POST", "api/view-log/", log, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a view log entry.
func (s *ViewLogService) Update(ctx context.Context, log *ViewLog) (*ViewLog, error) {
	var result ViewLog
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/view-log/%d/", log.ID), log, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a view log entry.
func (s *ViewLogService) Patch(ctx context.Context, log *ViewLog) (*ViewLog, error) {
	var result ViewLog
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/view-log/%d/", log.ID), log, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a view log entry.
func (s *ViewLogService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/view-log/%d/", id), nil, nil)
}

// UserFileService provides access to User File API endpoints.
type UserFileService struct {
	exec executor.Executor
}

// NewUserFileService creates a new UserFileService.
func NewUserFileService(e executor.Executor) *UserFileService {
	return &UserFileService{exec: e}
}

// List returns a paginated list of user files.
func (s *UserFileService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[UserFile], error) {
	path := "api/user-file/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[UserFile]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single user file by ID.
func (s *UserFileService) Get(ctx context.Context, id int) (*UserFile, error) {
	var file UserFile
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/user-file/%d/", id), nil, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// Create creates a new user file entry.
func (s *UserFileService) Create(ctx context.Context, file *UserFile) (*UserFile, error) {
	var result UserFile
	if err := s.exec.DoJSON(ctx, "POST", "api/user-file/", file, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a user file entry.
func (s *UserFileService) Update(ctx context.Context, file *UserFile) (*UserFile, error) {
	var result UserFile
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/user-file/%d/", file.ID), file, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a user file entry.
func (s *UserFileService) Patch(ctx context.Context, file *UserFile) (*UserFile, error) {
	var result UserFile
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/user-file/%d/", file.ID), file, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a user file entry.
func (s *UserFileService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/user-file/%d/", id), nil, nil)
}

// AutomationService provides access to Automation API endpoints.
type AutomationService struct {
	exec executor.Executor
}

// NewAutomationService creates a new AutomationService.
func NewAutomationService(e executor.Executor) *AutomationService {
	return &AutomationService{exec: e}
}

// List returns a paginated list of automations.
func (s *AutomationService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[Automation], error) {
	path := "api/automation/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Automation]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single automation by ID.
func (s *AutomationService) Get(ctx context.Context, id int) (*Automation, error) {
	var auto Automation
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/automation/%d/", id), nil, &auto); err != nil {
		return nil, err
	}
	return &auto, nil
}

// Create creates a new automation.
func (s *AutomationService) Create(ctx context.Context, auto *Automation) (*Automation, error) {
	var result Automation
	if err := s.exec.DoJSON(ctx, "POST", "api/automation/", auto, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces an automation.
func (s *AutomationService) Update(ctx context.Context, auto *Automation) (*Automation, error) {
	var result Automation
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/automation/%d/", auto.ID), auto, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on an automation.
func (s *AutomationService) Patch(ctx context.Context, auto *Automation) (*Automation, error) {
	var result Automation
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/automation/%d/", auto.ID), auto, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes an automation.
func (s *AutomationService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/automation/%d/", id), nil, nil)
}

// CustomFilterService provides access to Custom Filter API endpoints.
type CustomFilterService struct {
	exec executor.Executor
}

// NewCustomFilterService creates a new CustomFilterService.
func NewCustomFilterService(e executor.Executor) *CustomFilterService {
	return &CustomFilterService{exec: e}
}

// List returns a paginated list of custom filters.
func (s *CustomFilterService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[CustomFilter], error) {
	path := "api/custom-filter/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[CustomFilter]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single custom filter by ID.
func (s *CustomFilterService) Get(ctx context.Context, id int) (*CustomFilter, error) {
	var cf CustomFilter
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/custom-filter/%d/", id), nil, &cf); err != nil {
		return nil, err
	}
	return &cf, nil
}

// Create creates a new custom filter.
func (s *CustomFilterService) Create(ctx context.Context, cf *CustomFilter) (*CustomFilter, error) {
	var result CustomFilter
	if err := s.exec.DoJSON(ctx, "POST", "api/custom-filter/", cf, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a custom filter.
func (s *CustomFilterService) Update(ctx context.Context, cf *CustomFilter) (*CustomFilter, error) {
	var result CustomFilter
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/custom-filter/%d/", cf.ID), cf, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a custom filter.
func (s *CustomFilterService) Patch(ctx context.Context, cf *CustomFilter) (*CustomFilter, error) {
	var result CustomFilter
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/custom-filter/%d/", cf.ID), cf, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a custom filter.
func (s *CustomFilterService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/custom-filter/%d/", id), nil, nil)
}

// ConnectorConfigService provides access to Connector Config API endpoints.
type ConnectorConfigService struct {
	exec executor.Executor
}

// NewConnectorConfigService creates a new ConnectorConfigService.
func NewConnectorConfigService(e executor.Executor) *ConnectorConfigService {
	return &ConnectorConfigService{exec: e}
}

// List returns a paginated list of connector configurations.
func (s *ConnectorConfigService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[ConnectorConfig], error) {
	path := "api/connector-config/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[ConnectorConfig]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single connector configuration by ID.
func (s *ConnectorConfigService) Get(ctx context.Context, id int) (*ConnectorConfig, error) {
	var cc ConnectorConfig
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/connector-config/%d/", id), nil, &cc); err != nil {
		return nil, err
	}
	return &cc, nil
}

// Create creates a new connector configuration.
func (s *ConnectorConfigService) Create(ctx context.Context, cc *ConnectorConfig) (*ConnectorConfig, error) {
	var result ConnectorConfig
	if err := s.exec.DoJSON(ctx, "POST", "api/connector-config/", cc, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a connector configuration.
func (s *ConnectorConfigService) Update(ctx context.Context, cc *ConnectorConfig) (*ConnectorConfig, error) {
	var result ConnectorConfig
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/connector-config/%d/", cc.ID), cc, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a connector configuration.
func (s *ConnectorConfigService) Patch(ctx context.Context, cc *ConnectorConfig) (*ConnectorConfig, error) {
	var result ConnectorConfig
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/connector-config/%d/", cc.ID), cc, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a connector configuration.
func (s *ConnectorConfigService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/connector-config/%d/", id), nil, nil)
}

// SearchFieldsService provides access to Search Fields API endpoints.
type SearchFieldsService struct {
	exec executor.Executor
}

// NewSearchFieldsService creates a new SearchFieldsService.
func NewSearchFieldsService(e executor.Executor) *SearchFieldsService {
	return &SearchFieldsService{exec: e}
}

// List returns the available search fields. The endpoint responds with a
// flat array, not a paginated envelope.
func (s *SearchFieldsService) List(ctx context.Context, opts *pagination.ListOptions) ([]SearchField, error) {
	path := "api/search-fields/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var result []SearchField
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Get retrieves a single search field by ID.
func (s *SearchFieldsService) Get(ctx context.Context, id int) (*SearchField, error) {
	var sf SearchField
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/search-fields/%d/", id), nil, &sf); err != nil {
		return nil, err
	}
	return &sf, nil
}

// SearchPreferenceService provides access to Search Preference API endpoints.
type SearchPreferenceService struct {
	exec executor.Executor
}

// NewSearchPreferenceService creates a new SearchPreferenceService.
func NewSearchPreferenceService(e executor.Executor) *SearchPreferenceService {
	return &SearchPreferenceService{exec: e}
}

// List returns the search preferences. The endpoint responds with a flat
// array, not a paginated envelope.
func (s *SearchPreferenceService) List(ctx context.Context, opts *pagination.ListOptions) ([]SearchPreference, error) {
	path := "api/search-preference/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var result []SearchPreference
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Get retrieves a single search preference by ID.
func (s *SearchPreferenceService) Get(ctx context.Context, id int) (*SearchPreference, error) {
	var sp SearchPreference
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/search-preference/%d/", id), nil, &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// Patch performs a partial update on search preferences.
func (s *SearchPreferenceService) Patch(ctx context.Context, sp *SearchPreference) (*SearchPreference, error) {
	var result SearchPreference
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/search-preference/%d/", sp.User.ID), sp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AiProviderService provides access to AI Provider API endpoints.
type AiProviderService struct {
	exec executor.Executor
}

// NewAiProviderService creates a new AiProviderService.
func NewAiProviderService(e executor.Executor) *AiProviderService {
	return &AiProviderService{exec: e}
}

// List returns a paginated list of AI providers.
func (s *AiProviderService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[AiProvider], error) {
	path := "api/ai-provider/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[AiProvider]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single AI provider by ID.
func (s *AiProviderService) Get(ctx context.Context, id int) (*AiProvider, error) {
	var ap AiProvider
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/ai-provider/%d/", id), nil, &ap); err != nil {
		return nil, err
	}
	return &ap, nil
}

// Create creates a new AI provider.
func (s *AiProviderService) Create(ctx context.Context, req *AiProviderCreate) (*AiProvider, error) {
	var result AiProvider
	if err := s.exec.DoJSON(ctx, "POST", "api/ai-provider/", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces an AI provider.
func (s *AiProviderService) Update(ctx context.Context, ap *AiProvider) (*AiProvider, error) {
	var result AiProvider
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/ai-provider/%d/", ap.ID), ap, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on an AI provider.
func (s *AiProviderService) Patch(ctx context.Context, ap *AiProvider) (*AiProvider, error) {
	var result AiProvider
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/ai-provider/%d/", ap.ID), ap, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes an AI provider.
func (s *AiProviderService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/ai-provider/%d/", id), nil, nil)
}

// AiLogService provides access to AI Log API endpoints.
type AiLogService struct {
	exec executor.Executor
}

// NewAiLogService creates a new AiLogService.
func NewAiLogService(e executor.Executor) *AiLogService {
	return &AiLogService{exec: e}
}

// List returns a paginated list of AI logs.
func (s *AiLogService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[AiLog], error) {
	path := "api/ai-log/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[AiLog]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single AI log by ID.
func (s *AiLogService) Get(ctx context.Context, id int) (*AiLog, error) {
	var log AiLog
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/ai-log/%d/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// LocalizationService provides access to Localization API endpoints.
type LocalizationService struct {
	exec executor.Executor
}

// NewLocalizationService creates a new LocalizationService.
func NewLocalizationService(e executor.Executor) *LocalizationService {
	return &LocalizationService{exec: e}
}

// List returns the list of available localizations.
func (s *LocalizationService) List(ctx context.Context) ([]Localization, error) {
	var result []Localization
	if err := s.exec.DoJSON(ctx, "GET", "api/localization/", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ServerSettingsService provides access to Server Settings API endpoints.
type ServerSettingsService struct {
	exec executor.Executor
}

// NewServerSettingsService creates a new ServerSettingsService.
func NewServerSettingsService(e executor.Executor) *ServerSettingsService {
	return &ServerSettingsService{exec: e}
}

// Settings retrieves the current server settings.
func (s *ServerSettingsService) Settings(ctx context.Context) (*ServerSettings, error) {
	var result ServerSettings
	if err := s.exec.DoJSON(ctx, "GET", "api/server-settings/settings/", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
