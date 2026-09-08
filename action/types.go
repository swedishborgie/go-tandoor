// Package action contains types for Tandoor action endpoints (recipe-from-source, ai-import, etc.).
package action

// RecipeFromSourceRequest is the request body for /api/recipe-from-source/.
type RecipeFromSourceRequest struct {
	URL         string `json:"url"`         // URL to scrape (optional)
	Data        string `json:"data"`        // HTML/JSON data to parse (optional)
	Bookmarklet *int   `json:"bookmarklet"` // BookmarkletImport ID to use (overrides URL/Data)
}

// SourceImportStep represents a recipe step in the parsed response.
type SourceImportStep struct {
	Instruction          string `json:"instruction"`
	Ingredients          any    `json:"ingredients"` // list of parsed ingredients
	ShowIngredientsTable bool   `json:"show_ingredients_table"`
}

// SourceImportKeyword represents a keyword in the parsed response.
type SourceImportKeyword struct {
	ID            *int   `json:"id"`
	Label         string `json:"label"`
	Name          string `json:"name"`
	ImportKeyword bool   `json:"import_keyword"`
}

// SourceImportRecipe is the recipe data returned by recipe-from-source / ai-import.
type SourceImportRecipe struct {
	Steps        []SourceImportStep    `json:"steps"`
	Internal     bool                  `json:"internal"`
	SourceURL    string                `json:"source_url"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Servings     int                   `json:"servings"`
	ServingsText string                `json:"servings_text"`
	WorkingTime  int                   `json:"working_time"`
	WaitingTime  int                   `json:"waiting_time"`
	ImageURL     string                `json:"image_url"`
	Keywords     []SourceImportKeyword `json:"keywords"`
	Properties   []any                 `json:"properties"`
}

// SourceImportDuplicate represents a potential duplicate recipe.
type SourceImportDuplicate struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RecipeFromSourceResponse is the response from /api/recipe-from-source/ and /api/ai-import/.
type RecipeFromSourceResponse struct {
	Recipe     *SourceImportRecipe     `json:"recipe"`
	RecipeID   *int                    `json:"recipe_id"`
	Images     []string                `json:"images"`
	Error      bool                    `json:"error"`
	Msg        string                  `json:"msg"`
	Duplicates []SourceImportDuplicate `json:"duplicates"`
}

// AiImportRequest is the request body for /api/ai-import/.
type AiImportRequest struct {
	AIProviderID int    `json:"ai_provider_id"`
	File         string `json:"file,omitempty"`      // multipart file (image/PDF)
	Text         string `json:"text,omitempty"`      // raw text
	RecipeID     *int   `json:"recipe_id,omitempty"` // existing recipe with file
}

// IngredientParserRequest is the request body for /api/ingredient-parser/.
type IngredientParserRequest struct {
	Ingredient  string   `json:"ingredient"`  // single ingredient string
	Ingredients []string `json:"ingredients"` // list of ingredient strings
}

// ParsedIngredient represents a single parsed ingredient.
type ParsedIngredient struct {
	Amount       *float64    `json:"amount"`
	Food         *ParsedFood `json:"food"`
	Unit         *ParsedUnit `json:"unit"`
	Note         string      `json:"note"`
	OriginalText string      `json:"original_text"`
	Order        *int        `json:"order"`
	Checked      bool        `json:"checked"`
}

// ParsedFood represents a food reference in a parsed ingredient.
type ParsedFood struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ParsedUnit represents a unit reference in a parsed ingredient.
type ParsedUnit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// IngredientParserResponse is the response from /api/ingredient-parser/.
type IngredientParserResponse struct {
	Ingredient  *ParsedIngredient  `json:"ingredient"`
	Ingredients []ParsedIngredient `json:"ingredients"`
	Error       bool               `json:"error"`
	Msg         string             `json:"msg"`
}

// FdcFood represents a food item from the USDA FDC API.
type FdcFood struct {
	FDCID       int    `json:"fdcId"`
	Description string `json:"description"`
	DataType    string `json:"dataType"`
}

// FdcQueryResponse is the response from /api/fdc-search/.
type FdcQueryResponse struct {
	TotalHits   int       `json:"totalHits"`
	CurrentPage int       `json:"currentPage"`
	TotalPages  int       `json:"totalPages"`
	Foods       []FdcFood `json:"foods"`
}

// ImportOpenDataRequest is the request body for /api/import-open-data/ POST.
type ImportOpenDataRequest struct {
	SelectedVersion   string   `json:"selected_version"`
	SelectedDataTypes []string `json:"selected_datatypes"`
	UpdateExisting    bool     `json:"update_existing"`
	UseMetric         bool     `json:"use_metric"`
}

// ImportOpenDataResult is the per-datatype result of an open data import.
type ImportOpenDataResult struct {
	TotalCreated   int `json:"total_created"`
	TotalUpdated   int `json:"total_updated"`
	TotalUntouched int `json:"total_untouched"`
	TotalErrored   int `json:"total_errored"`
}

// ImportOpenDataResponse is the full response from import-open-data POST.
type ImportOpenDataResponse map[string]ImportOpenDataResult

// ImportOpenDataMetadata describes available open data versions/datatypes.
type ImportOpenDataMetadata struct {
	Versions  []string       `json:"versions"`
	DataTypes []string       `json:"datatypes"`
	Base      DataTypeCounts `json:"base"`
	En        DataTypeCounts `json:"en"`
	De        DataTypeCounts `json:"de"`
	Es        DataTypeCounts `json:"es"`
	Fr        DataTypeCounts `json:"fr"`
	It        DataTypeCounts `json:"it"`
	// other locales follow same shape
}

// DataTypeCounts describes counts per datatype in open data.
type DataTypeCounts struct {
	Food       int `json:"food"`
	Unit       int `json:"unit"`
	Category   int `json:"category"`
	Property   int `json:"property"`
	Store      int `json:"store"`
	Conversion int `json:"conversion"`
}

// ShareLinkResponse is the response from /api/share-link/<pk>.
type ShareLinkResponse struct {
	PK    int    `json:"pk"`
	Share string `json:"share"`
	Link  string `json:"link"`
}

// AppImportResult is the response from /api/import/.
type AppImportResult struct {
	ImportID int `json:"import_id"`
}
