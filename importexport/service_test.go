package importexport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

// RecipeImportService tests

func TestRecipeImportService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/recipe-import/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Test Recipe","file_uid":"abc123"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewRecipeImportService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "Test Recipe", page.Results[0].Name)
}

func TestRecipeImportService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/recipe-import/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Test Recipe","file_path":"/path/to/file"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	ri, err := NewRecipeImportService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Test Recipe", ri.Name)
}

func TestRecipeImportService_ImportRecipe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/recipe-import/1/import_recipe/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":100,"name":"Converted Recipe"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	result, err := NewRecipeImportService(mock).ImportRecipe(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 100, result.ID)
}

func TestRecipeImportService_ImportAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/recipe-import/import_all/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"msg":"ok"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	result, err := NewRecipeImportService(mock).ImportAll(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRecipeImportService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/recipe-import/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewRecipeImportService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// ImportLogService tests

func TestImportLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/import-log/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"type":"JSON","running":false,"imported_recipes":10}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewImportLogService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "JSON", page.Results[0].Type)
	assert.Equal(t, 10, page.Results[0].ImportedRecipes)
}

func TestImportLogService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/import-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"type":"CSV","running":true,"total_recipes":5}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewImportLogService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "CSV", log.Type)
	assert.True(t, log.Running)
}

// ExportLogService tests

func TestExportLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/export-log/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"type":"JSON","running":false,"exported_recipes":50}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewExportLogService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "JSON", page.Results[0].Type)
}

func TestExportLogService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/export-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"type":"CSV","running":false,"total_recipes":100,"exported_recipes":100}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewExportLogService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "CSV", log.Type)
	assert.Equal(t, 100, log.ExportedRecipes)
}

// BookmarkletImportService tests

func TestBookmarkletImportService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/bookmarklet-import/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"url":"https://example.com/recipe/1"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewBookmarkletImportService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
}

func TestBookmarkletImportService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/bookmarklet-import/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"url":"https://example.com/recipe/1","created_by":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	bi, err := NewBookmarkletImportService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Contains(t, bi.URL, "example.com")
}

func TestBookmarkletImportService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"url":"https://example.com/recipe/10"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	bi, err := NewBookmarkletImportService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Contains(t, bi.URL, "example.com")
}

func TestBookmarkletImportService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/bookmarklet-import/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewBookmarkletImportService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// SyncService tests

func TestSyncService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/sync/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"path":"/recipes","active":true}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewSyncService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.True(t, page.Results[0].Active)
}

func TestSyncService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/sync/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"path":"/dropbox/recipes","active":true,"last_checked":""}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sync, err := NewSyncService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "/dropbox/recipes", sync.Path)
}

func TestSyncService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"path":"/sync/path","active":true}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sync, err := NewSyncService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "/sync/path", sync.Path)
}

func TestSyncService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"path":"/updated/path","active":false}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sync, err := NewSyncService(mock).Update(context.Background(), &Sync{ID: 1})
	require.NoError(t, err)
	assert.False(t, sync.Active)
}

func TestSyncService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sync, err := NewSyncService(mock).Patch(context.Background(), &Sync{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, sync.ID)
}

func TestSyncService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/sync/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewSyncService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestSyncService_QuerySyncedFolder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/sync/1/query_synced_folder/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":5,"status":"success","msg":"synced 3 files"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewSyncService(mock).QuerySyncedFolder(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "success", log.Status)
}

// SyncLogService tests

func TestSyncLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/sync-log/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"status":"success","msg":"3 files synced"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewSyncLogService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "success", page.Results[0].Status)
}

func TestSyncLogService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/sync-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"status":"error","msg":"connection timeout"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewSyncLogService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "error", log.Status)
}

// AppExportService tests

func TestAppExportService_Export(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/export/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":10,"type":"JSON","running":true,"total_recipes":5}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewAppExportService(mock).Export(context.Background(), &AppExportRequest{
		Type: "JSON",
		All:  true,
	})
	require.NoError(t, err)
	assert.Equal(t, "JSON", log.Type)
	assert.True(t, log.Running)
}

func TestAppExportService_DownloadExportFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/export-file/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"recipes":[{}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewAppExportService(mock).DownloadExportFile(context.Background(), 1, "/tmp/test_export.json")
	assert.NoError(t, err)
}
