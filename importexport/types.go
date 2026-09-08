// Package importexport contains types for Tandoor Import/Export resources.
package importexport

// RecipeImport represents a recipe that was imported via storage sync.
type RecipeImport struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	FileUID   string `json:"file_uid"`
	FilePath  string `json:"file_path"`
	CreatedAt string `json:"created_at"`
	Storage   any    `json:"storage"` // Storage object (nested)
}

// ImportLog records the result of an import operation (file-based or bookmarklet).
type ImportLog struct {
	ID              int    `json:"id"`
	Type            string `json:"type"` // e.g. "JSON", "CSV", "PDF"
	Msg             string `json:"msg"`
	Running         bool   `json:"running"`
	TotalRecipes    int    `json:"total_recipes"`
	ImportedRecipes int    `json:"imported_recipes"`
	CreatedBy       int    `json:"created_by"`
	CreatedAt       string `json:"created_at"`
	Keyword         any    `json:"keyword"` // Keyword object (nested, optional)
}

// ExportLog records the result of an export operation.
type ExportLog struct {
	ID                 int    `json:"id"`
	Type               string `json:"type"` // e.g. "PDF", "CSV"
	Msg                string `json:"msg"`
	Running            bool   `json:"running"`
	TotalRecipes       int    `json:"total_recipes"`
	ExportedRecipes    int    `json:"exported_recipes"`
	CacheDuration      string `json:"cache_duration"`
	PossiblyNotExpired bool   `json:"possibly_not_expired"`
	CreatedBy          int    `json:"created_by"`
	CreatedAt          string `json:"created_at"`
}

// BookmarkletImport stores a recipe bookmarked via the browser bookmarklet.
type BookmarkletImport struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	HTML      string `json:"html"`
	CreatedBy int    `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

// Sync represents a file sync configuration (Dropbox, Nextcloud, Local).
type Sync struct {
	ID          int    `json:"id"`
	Storage     any    `json:"storage"` // Storage object (nested)
	Path        string `json:"path"`
	Active      bool   `json:"active"`
	LastChecked string `json:"last_checked"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// SyncLog records the result of a sync operation.
type SyncLog struct {
	ID        int    `json:"id"`
	Sync      any    `json:"sync"` // Sync object (nested, read-only)
	Status    string `json:"status"`
	Msg       string `json:"msg"`
	CreatedAt string `json:"created_at"`
}

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
	// ... other locales follow same shape
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

// ImportOpenDataResult is the per-datatype result of an open data import.
type ImportOpenDataResult struct {
	TotalCreated   int `json:"total_created"`
	TotalUpdated   int `json:"total_updated"`
	TotalUntouched int `json:"total_untouched"`
	TotalErrored   int `json:"total_errored"`
}

// ImportOpenDataResponse is the full response from import-open-data POST.
type ImportOpenDataResponse map[string]ImportOpenDataResult

// AppImportResult is the response from the app import endpoint.
type AppImportResult struct {
	ImportID int `json:"import_id"`
}

// AppExportRequest is the request body for the app export endpoint.
type AppExportRequest struct {
	Type         string           `json:"type"` // e.g. "JSON", "CSV"
	All          bool             `json:"all"`
	Recipes      []map[string]any `json:"recipes"`       // [{"id": 1}, ...]
	CustomFilter *map[string]any  `json:"custom_filter"` // optional filter reference
}
