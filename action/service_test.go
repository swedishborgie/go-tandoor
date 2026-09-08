package action

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

// RecipeFromSourceService tests

func TestRecipeFromSourceService_Import(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/recipe-from-source/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"recipe":{"id":1,"name":"Pasta"},"images":["https://example.com/img.jpg"],"error":false,"duplicates":[]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewRecipeFromSourceService(mock).Import(context.Background(), &RecipeFromSourceRequest{
		URL: "https://example.com/recipe/1",
	})
	require.NoError(t, err)
	assert.False(t, resp.Error)
	assert.Equal(t, "Pasta", resp.Recipe.Name)
}

func TestRecipeFromSourceService_Import_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"error":true,"msg":"No usable data could be found."}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewRecipeFromSourceService(mock).Import(context.Background(), &RecipeFromSourceRequest{
		URL: "https://invalid.example.com",
	})
	require.NoError(t, err)
	assert.True(t, resp.Error)
}

// AiImportService tests

func TestAiImportService_Import(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/ai-import/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"recipe":{"id":2,"name":"AI Recipe"},"error":false,"duplicates":[]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewAiImportService(mock).Import(context.Background(), &AiImportRequest{
		AIProviderID: 1,
		Text:         "ingredients: 2 cups flour, 3 eggs",
	})
	require.NoError(t, err)
	assert.False(t, resp.Error)
	assert.Equal(t, "AI Recipe", resp.Recipe.Name)
}

// AiStepSortService tests

func TestAiStepSortService_Sort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/ai-step-sort/", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "provider=1")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"steps":[{"instruction":"Mix dry ingredients"},{"instruction":"Add wet ingredients"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	result, err := NewAiStepSortService(mock).Sort(context.Background(), map[string]any{
		"name":  "Test Recipe",
		"steps": []any{"Mix everything"},
	}, 1)
	require.NoError(t, err)
	steps := result["steps"].([]any)
	assert.Len(t, steps, 2)
}

// IngredientParserService tests

func TestIngredientParserService_Parse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/ingredient-parser/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ingredient":{"amount":2.0,"note":"organic"},"ingredients":[]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewIngredientParserService(mock).Parse(context.Background(), &IngredientParserRequest{
		Ingredient: "2 cups flour",
	})
	require.NoError(t, err)
	assert.InEpsilon(t, 2.0, *resp.Ingredient.Amount, 1e-9)
	assert.Equal(t, "organic", resp.Ingredient.Note)
}

// FdcSearchService tests

func TestFdcSearchService_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/fdc-search/", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "query=chicken")
		assert.Contains(t, r.URL.RawQuery, "dataType=Foundation")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"totalHits":1,"currentPage":1,"totalPages":1,"foods":[{"fdcId":53393,"description":"Chicken breast","dataType":"Foundation"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewFdcSearchService(mock).Search(context.Background(), "chicken", []string{"Foundation"})
	require.NoError(t, err)
	assert.Equal(t, 1, resp.TotalHits)
	assert.Equal(t, 53393, resp.Foods[0].FDCID)
	assert.Equal(t, "Chicken breast", resp.Foods[0].Description)
}

// ImportOpenDataService tests

func TestImportOpenDataService_GetMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/import-open-data/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"versions":["v1"],"datatypes":["food","unit"],"base":{"food":100,"unit":50}}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewImportOpenDataService(mock).GetMetadata(context.Background())
	require.NoError(t, err)
	assert.Contains(t, resp.Versions, "v1")
	assert.Equal(t, 100, resp.Base.Food)
}

func TestImportOpenDataService_Import(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/import-open-data/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"food":{"total_created":10,"total_updated":5,"total_untouched":90,"total_errored":0}}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewImportOpenDataService(mock).Import(context.Background(), &ImportOpenDataRequest{
		SelectedVersion:   "en.json",
		SelectedDataTypes: []string{"food"},
		UpdateExisting:    true,
		UseMetric:         true,
	})
	require.NoError(t, err)
	foodResult := (*resp)["food"]
	assert.Equal(t, 10, foodResult.TotalCreated)
	assert.Equal(t, 5, foodResult.TotalUpdated)
}

// ShareLinkService tests

func TestShareLinkService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/share-link/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"pk":1,"share":"abcd-1234","link":"https://example.com/recipe/1/?share=abcd-1234"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	resp, err := NewShareLinkService(mock).Create(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.PK)
	assert.Equal(t, "abcd-1234", resp.Share)
	assert.Contains(t, resp.Link, "example.com")
}
