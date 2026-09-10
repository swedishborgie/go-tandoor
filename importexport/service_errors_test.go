package importexport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

func errExecutor(t *testing.T) *testutil.MockExecutor {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"Not found."}`))
	}))
	t.Cleanup(server.Close)
	return testutil.NewMockExecutor(server.URL)
}

// TestImportExportErrorBranches drives every import/export service method
// against a 404 server to cover the error branches.
func TestImportExportErrorBranches(t *testing.T) {
	e := errExecutor(t)
	ctx := context.Background()

	ri := NewRecipeImportService(e)
	if _, err := ri.List(ctx, &RecipeImportListOptions{}); err == nil {
		t.Error("RecipeImport List: expected error")
	}
	if _, err := ri.Get(ctx, 1); err == nil {
		t.Error("RecipeImport Get: expected error")
	}
	if _, err := ri.ImportRecipe(ctx, 1); err == nil {
		t.Error("RecipeImport ImportRecipe: expected error")
	}
	if _, err := ri.ImportAll(ctx); err == nil {
		t.Error("RecipeImport ImportAll: expected error")
	}
	if err := ri.Delete(ctx, 1); err == nil {
		t.Error("RecipeImport Delete: expected error")
	}

	il := NewImportLogService(e)
	if _, err := il.List(ctx, &ImportLogListOptions{}); err == nil {
		t.Error("ImportLog List: expected error")
	}
	if _, err := il.Get(ctx, 1); err == nil {
		t.Error("ImportLog Get: expected error")
	}

	el := NewExportLogService(e)
	if _, err := el.List(ctx, &ExportLogListOptions{}); err == nil {
		t.Error("ExportLog List: expected error")
	}
	if _, err := el.Get(ctx, 1); err == nil {
		t.Error("ExportLog Get: expected error")
	}

	bi := NewBookmarkletImportService(e)
	if _, err := bi.List(ctx, &BookmarkletImportListOptions{}); err == nil {
		t.Error("BookmarkletImport List: expected error")
	}
	if _, err := bi.Get(ctx, 1); err == nil {
		t.Error("BookmarkletImport Get: expected error")
	}
	if _, err := bi.Create(ctx, &BookmarkletImport{}); err == nil {
		t.Error("BookmarkletImport Create: expected error")
	}
	if err := bi.Delete(ctx, 1); err == nil {
		t.Error("BookmarkletImport Delete: expected error")
	}

	s := NewSyncService(e)
	if _, err := s.List(ctx, &SyncListOptions{}); err == nil {
		t.Error("Sync List: expected error")
	}
	if _, err := s.Get(ctx, 1); err == nil {
		t.Error("Sync Get: expected error")
	}
	if _, err := s.Create(ctx, &Sync{}); err == nil {
		t.Error("Sync Create: expected error")
	}
	if _, err := s.Update(ctx, &Sync{ID: 1}); err == nil {
		t.Error("Sync Update: expected error")
	}
	if _, err := s.Patch(ctx, &Sync{ID: 1}); err == nil {
		t.Error("Sync Patch: expected error")
	}
	if err := s.Delete(ctx, 1); err == nil {
		t.Error("Sync Delete: expected error")
	}
	if _, err := s.QuerySyncedFolder(ctx, 1); err == nil {
		t.Error("Sync QuerySyncedFolder: expected error")
	}

	sl := NewSyncLogService(e)
	if _, err := sl.List(ctx, &SyncLogListOptions{}); err == nil {
		t.Error("SyncLog List: expected error")
	}
	if _, err := sl.Get(ctx, 1); err == nil {
		t.Error("SyncLog Get: expected error")
	}

	ae := NewAppExportService(e)
	if _, err := ae.Export(ctx, &AppExportRequest{}); err == nil {
		t.Error("AppExport Export: expected error")
	}
	// DownloadExportFile uses DoRaw (no status check), so the error branch
	// only fires on transport failure — point at a closed port.
	ae2 := NewAppExportService(testutil.NewMockExecutor("http://127.0.0.1:1"))
	if err := ae2.DownloadExportFile(ctx, 1, t.TempDir()+"/out.json"); err == nil {
		t.Error("AppExport DownloadExportFile: expected error")
	}
}
