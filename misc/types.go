// Package misc contains types for Tandoor misc resources (CookLog, ViewLog, Automation, etc.).
package misc

import (
	"github.com/swedishborgie/go-tandoor/space"
)

// CookLog represents a log entry for when a recipe was cooked.
//
// Endpoints: GET/POST api/cook-log/ GET/PUT/PATCH/DELETE api/cook-log/<id>/
type CookLog struct {
	ID        int         `json:"id"`
	Recipe    *int        `json:"recipe"`
	Servings  *int        `json:"servings,omitempty"`
	Rating    *int        `json:"rating,omitempty"`
	Comment   *string     `json:"comment,omitempty"`
	CreatedBy *space.User `json:"created_by,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
	UpdatedAt string      `json:"updated_at,omitempty"`
}

// ViewLog represents a log entry for when a recipe was viewed.
//
// Endpoints: GET/POST api/view-log/ GET/PUT/PATCH/DELETE api/view-log/<id>/
// Note: create merges duplicate views within 5 minutes.
type ViewLog struct {
	ID        int    `json:"id"`
	Recipe    *int   `json:"recipe"`
	CreatedBy *int   `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

// UserFile represents a file uploaded by a user.
//
// Endpoints: GET/POST api/user-file/ GET/PUT/PATCH/DELETE api/user-file/<id>/
// Note: file field is write-only (upload), file_download/preview are read-only.
type UserFile struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	File         string      `json:"file"`
	FileSizeKB   int         `json:"file_size_kb"`
	CreatedAt    string      `json:"created_at"`
	CreatedBy    *space.User `json:"created_by"`
	FileDownload string      `json:"file_download"`
	Preview      string      `json:"preview"`
}

// Automation represents a rule for automating recipe ingestion.
//
// Endpoints: GET/POST api/automation/ GET/PUT/PATCH/DELETE api/automation/<id>/
// Types: FOOD_ALIAS, UNIT_ALIAS, KEYWORD_ALIAS, DESCRIPTION_REPLACE,
//
//	INSTRUCTION_REPLACE, NEVER_UNIT, TRANSPOSE_WORDS, FOOD_REPLACE,
//	UNIT_REPLACE, NAME_REPLACE
type Automation struct {
	ID          int         `json:"id"`
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Param1      string      `json:"param_1"`
	Param2      string      `json:"param_2"`
	Param3      string      `json:"param_3"`
	Order       int         `json:"order"`
	Disabled    bool        `json:"disabled"`
	CreatedBy   *space.User `json:"created_by"`
}

// CustomFilter represents a saved filter for recipes, foods, or keywords.
//
// Endpoints: GET/POST api/custom-filter/ GET/PUT/PATCH/DELETE api/custom-filter/<id>/
type CustomFilter struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Search    string       `json:"search"`
	Shared    []space.User `json:"shared"`
	CreatedBy *space.User  `json:"created_by"`
}

// ConnectorConfig represents a third-party connector configuration (e.g., HomeAssistant).
//
// Endpoints: GET/POST api/connector-config/ GET/PUT/PATCH/DELETE api/connector-config/<id>/
type ConnectorConfig struct {
	ID                                int         `json:"id"`
	Name                              string      `json:"name"`
	Type                              string      `json:"type"`
	URL                               string      `json:"url"`
	Token                             string      `json:"token"`
	TodoEntity                        string      `json:"todo_entity"`
	Enabled                           bool        `json:"enabled"`
	OnShoppingListEntryCreatedEnabled bool        `json:"on_shopping_list_entry_created_enabled"`
	OnShoppingListEntryUpdatedEnabled bool        `json:"on_shopping_list_entry_updated_enabled"`
	OnShoppingListEntryDeletedEnabled bool        `json:"on_shopping_list_entry_deleted_enabled"`
	SupportsDescriptionField          bool        `json:"supports_description_field"`
	CreatedBy                         *space.User `json:"created_by"`
}

// SearchField represents an available field for searching.
//
// Endpoints: GET api/search-fields/ GET api/search-fields/<id>/
// Note: create is blocked, update is no-op (read-only effectively).
type SearchField struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Field string `json:"field"`
}

// SearchPreference represents user-specific search configuration.
//
// Endpoints: GET/POST api/search-preference/ GET/PUT/PATCH/DELETE api/search-preference/<id>/
// Note: create is blocked, only patch/update allowed.
type SearchPreference struct {
	User             *space.User   `json:"user"`
	Search           string        `json:"search"`
	Lookup           bool          `json:"lookup"`
	Unaccent         []SearchField `json:"unaccent"`
	IContains        []SearchField `json:"icontains"`
	IStartsWith      []SearchField `json:"istartswith"`
	Trigram          []SearchField `json:"trigram"`
	Fulltext         []SearchField `json:"fulltext"`
	TrigramThreshold float64       `json:"trigram_threshold"`
}

// AiProvider represents an AI provider configuration for AI-powered features.
//
// Endpoints: GET/POST api/ai-provider/ GET/PUT/PATCH/DELETE api/ai-provider/<id>/
// Note: api_key is write-only (not returned on read).
type AiProvider struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Space         *int   `json:"space"`
	ModelName     string `json:"model_name"`
	URL           string `json:"url"`
	LogCreditCost bool   `json:"log_credit_cost"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// AiProviderCreate is the request body for creating an AI provider.
//
// API key is required on create.
type AiProviderCreate struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Space         *int   `json:"space,omitempty"`
	APIKey        string `json:"api_key"`
	ModelName     string `json:"model_name"`
	URL           string `json:"url,omitempty"`
	LogCreditCost *bool  `json:"log_credit_cost,omitempty"`
}

// AiLog represents a log entry for an AI operation.
//
// Endpoints: GET api/ai-log/ GET api/ai-log/<id>/ (read-only)
type AiLog struct {
	ID                 int         `json:"id"`
	AiProvider         *AiProvider `json:"ai_provider"`
	Function           string      `json:"function"`
	CreditCost         float64     `json:"credit_cost"`
	CreditsFromBalance bool        `json:"credits_from_balance"`
	InputTokens        int         `json:"input_tokens"`
	OutputTokens       int         `json:"output_tokens"`
	StartTime          string      `json:"start_time"`
	EndTime            string      `json:"end_time"`
	CreatedBy          *space.User `json:"created_by"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
}

// Localization represents an available language/locale.
//
// Endpoints: GET api/localization/ (list only, no detail endpoint)
type Localization struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// ServerSettings represents global server configuration.
//
// Endpoints: GET api/server-settings/settings/ (read-only)
type ServerSettings struct {
	ShoppingMinAutosyncInterval   string `json:"shopping_min_autosync_interval"`
	DisableExternalConnectors     bool   `json:"disable_external_connectors"`
	TermsURL                      string `json:"terms_url"`
	PrivacyURL                    string `json:"privacy_url"`
	ImprintURL                    string `json:"imprint_url"`
	Hosted                        bool   `json:"hosted"`
	Debug                         bool   `json:"debug"`
	Version                       string `json:"version"`
	UnauthenticatedThemeFromSpace int    `json:"unauthenticated_theme_from_space"`
	ForceThemeFromSpace           int    `json:"force_theme_from_space"`
	LogoColor32                   string `json:"logo_color_32"`
	LogoColor128                  string `json:"logo_color_128"`
	LogoColor144                  string `json:"logo_color_144"`
	LogoColor180                  string `json:"logo_color_180"`
	LogoColor192                  string `json:"logo_color_192"`
	LogoColor512                  string `json:"logo_color_512"`
	LogoColorSVG                  string `json:"logo_color_svg"`
	CustomSpaceTheme              string `json:"custom_space_theme"`
	NavLogo                       string `json:"nav_logo"`
	NavBGColor                    string `json:"nav_bg_color"`
}
