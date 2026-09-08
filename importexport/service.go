package importexport

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// RecipeImportService provides access to Recipe Import API endpoints.
type RecipeImportService struct {
	exec executor.Executor
}

// NewRecipeImportService creates a new RecipeImportService.
func NewRecipeImportService(e executor.Executor) *RecipeImportService {
	return &RecipeImportService{exec: e}
}

// List returns a paginated list of recipe imports.
func (s *RecipeImportService) List(ctx context.Context, opts *RecipeImportListOptions) (*pagination.Paginated[RecipeImport], error) {
	path := "api/recipe-import/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[RecipeImport]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single recipe import by ID.
func (s *RecipeImportService) Get(ctx context.Context, id int) (*RecipeImport, error) {
	var ri RecipeImport
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/recipe-import/%d/", id), nil, &ri); err != nil {
		return nil, err
	}
	return &ri, nil
}

// ImportRecipe triggers the import of a single recipe.
func (s *RecipeImportService) ImportRecipe(ctx context.Context, id int) (*RecipeImport, error) {
	var result RecipeImport
	if err := s.exec.DoJSON(ctx, "POST", fmt.Sprintf("api/recipe-import/%d/import_recipe/", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImportAll triggers the import of all pending recipes.
func (s *RecipeImportService) ImportAll(ctx context.Context) (*RecipeImport, error) {
	var result RecipeImport
	if err := s.exec.DoJSON(ctx, "POST", "api/recipe-import/import_all/", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a recipe import by ID.
func (s *RecipeImportService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/recipe-import/%d/", id), nil, nil)
}

// ImportLogService provides access to Import Log API endpoints.
type ImportLogService struct {
	exec executor.Executor
}

// NewImportLogService creates a new ImportLogService.
func NewImportLogService(e executor.Executor) *ImportLogService {
	return &ImportLogService{exec: e}
}

// List returns a paginated list of import logs.
func (s *ImportLogService) List(ctx context.Context, opts *ImportLogListOptions) (*pagination.Paginated[ImportLog], error) {
	path := "api/import-log/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[ImportLog]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single import log by ID.
func (s *ImportLogService) Get(ctx context.Context, id int) (*ImportLog, error) {
	var log ImportLog
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/import-log/%d/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// ExportLogService provides access to Export Log API endpoints.
type ExportLogService struct {
	exec executor.Executor
}

// NewExportLogService creates a new ExportLogService.
func NewExportLogService(e executor.Executor) *ExportLogService {
	return &ExportLogService{exec: e}
}

// List returns a paginated list of export logs.
func (s *ExportLogService) List(ctx context.Context, opts *ExportLogListOptions) (*pagination.Paginated[ExportLog], error) {
	path := "api/export-log/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[ExportLog]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single export log by ID.
func (s *ExportLogService) Get(ctx context.Context, id int) (*ExportLog, error) {
	var log ExportLog
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/export-log/%d/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// BookmarkletImportService provides access to Bookmarklet Import API endpoints.
type BookmarkletImportService struct {
	exec executor.Executor
}

// NewBookmarkletImportService creates a new BookmarkletImportService.
func NewBookmarkletImportService(e executor.Executor) *BookmarkletImportService {
	return &BookmarkletImportService{exec: e}
}

// List returns a paginated list of bookmarklet imports.
func (s *BookmarkletImportService) List(ctx context.Context, opts *BookmarkletImportListOptions) (*pagination.Paginated[BookmarkletImport], error) {
	path := "api/bookmarklet-import/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[BookmarkletImport]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single bookmarklet import by ID.
func (s *BookmarkletImportService) Get(ctx context.Context, id int) (*BookmarkletImport, error) {
	var bi BookmarkletImport
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/bookmarklet-import/%d/", id), nil, &bi); err != nil {
		return nil, err
	}
	return &bi, nil
}

// Create creates a new bookmarklet import.
func (s *BookmarkletImportService) Create(ctx context.Context, bi *BookmarkletImport) (*BookmarkletImport, error) {
	var result BookmarkletImport
	if err := s.exec.DoJSON(ctx, "POST", "api/bookmarklet-import/", bi, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a bookmarklet import by ID.
func (s *BookmarkletImportService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/bookmarklet-import/%d/", id), nil, nil)
}

// SyncService provides access to Sync API endpoints.
type SyncService struct {
	exec executor.Executor
}

// NewSyncService creates a new SyncService.
func NewSyncService(e executor.Executor) *SyncService {
	return &SyncService{exec: e}
}

// List returns a paginated list of syncs.
func (s *SyncService) List(ctx context.Context, opts *SyncListOptions) (*pagination.Paginated[Sync], error) {
	path := "api/sync/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Sync]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single sync by ID.
func (s *SyncService) Get(ctx context.Context, id int) (*Sync, error) {
	var sync Sync
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/sync/%d/", id), nil, &sync); err != nil {
		return nil, err
	}
	return &sync, nil
}

// Create creates a new sync.
func (s *SyncService) Create(ctx context.Context, sync *Sync) (*Sync, error) {
	var result Sync
	if err := s.exec.DoJSON(ctx, "POST", "api/sync/", sync, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a sync.
func (s *SyncService) Update(ctx context.Context, sync *Sync) (*Sync, error) {
	var result Sync
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/sync/%d/", sync.ID), sync, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a sync.
func (s *SyncService) Patch(ctx context.Context, sync *Sync) (*Sync, error) {
	var result Sync
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/sync/%d/", sync.ID), sync, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a sync by ID.
func (s *SyncService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/sync/%d/", id), nil, nil)
}

// QuerySyncedFolder queries a synced folder for changes.
func (s *SyncService) QuerySyncedFolder(ctx context.Context, id int) (*SyncLog, error) {
	var log SyncLog
	if err := s.exec.DoJSON(ctx, "POST", fmt.Sprintf("api/sync/%d/query_synced_folder/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// SyncLogService provides access to Sync Log API endpoints.
type SyncLogService struct {
	exec executor.Executor
}

// NewSyncLogService creates a new SyncLogService.
func NewSyncLogService(e executor.Executor) *SyncLogService {
	return &SyncLogService{exec: e}
}

// List returns a paginated list of sync logs.
func (s *SyncLogService) List(ctx context.Context, opts *SyncLogListOptions) (*pagination.Paginated[SyncLog], error) {
	path := "api/sync-log/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[SyncLog]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single sync log by ID.
func (s *SyncLogService) Get(ctx context.Context, id int) (*SyncLog, error) {
	var log SyncLog
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/sync-log/%d/", id), nil, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// AppExportService provides access to the app-level export endpoint.
type AppExportService struct {
	exec executor.Executor
}

// NewAppExportService creates a new AppExportService.
func NewAppExportService(e executor.Executor) *AppExportService {
	return &AppExportService{exec: e}
}

// Export creates an export job.
func (s *AppExportService) Export(ctx context.Context, req *AppExportRequest) (*ExportLog, error) {
	var log ExportLog
	if err := s.exec.DoJSON(ctx, "POST", "api/export/", req, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

// DownloadExportFile downloads an export file to the given path.
//
// Note: this uses raw HTTP (not DoJSON) since it downloads a file.
func (s *AppExportService) DownloadExportFile(ctx context.Context, id int, dst string) error {
	path := fmt.Sprintf("export-file/%d/", id)
	resp, err := s.exec.DoRaw(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
