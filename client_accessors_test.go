package tandoor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_Accessors exercises every service accessor once.
func TestClient_Accessors(t *testing.T) {
	c, err := NewClient("http://example.test")
	require.NoError(t, err)

	assert.NotNil(t, c.Auth())
	assert.NotNil(t, c.Foods())
	assert.NotNil(t, c.Ingredients())
	assert.NotNil(t, c.Keywords())
	assert.NotNil(t, c.Steps())
	assert.NotNil(t, c.Units())
	assert.NotNil(t, c.UnitConversions())
	assert.NotNil(t, c.Recipes())
	assert.NotNil(t, c.RecipeBooks())
	assert.NotNil(t, c.RecipeBookEntries())
	assert.NotNil(t, c.Properties())
	assert.NotNil(t, c.PropertyTypes())
	assert.NotNil(t, c.Inventory())
	assert.NotNil(t, c.InventoryLocations())
	assert.NotNil(t, c.InventoryEntries())
	assert.NotNil(t, c.InventoryLogs())
	assert.NotNil(t, c.Storages())
	assert.NotNil(t, c.Spaces())
	assert.NotNil(t, c.Users())
	assert.NotNil(t, c.Groups())
	assert.NotNil(t, c.Households())
	assert.NotNil(t, c.UserSpaces())
	assert.NotNil(t, c.UserPreferences())
	assert.NotNil(t, c.InviteLinks())
	assert.NotNil(t, c.MealPlans())
	assert.NotNil(t, c.MealTypes())
	assert.NotNil(t, c.AutoPlan())
	assert.NotNil(t, c.ShoppingLists())
	assert.NotNil(t, c.ShoppingEntries())
	assert.NotNil(t, c.ShoppingRecipes())
	assert.NotNil(t, c.Supermarkets())
	assert.NotNil(t, c.SupermarketCategories())
	assert.NotNil(t, c.SupermarketCategoryRelations())
	assert.NotNil(t, c.RecipeFromSource())
	assert.NotNil(t, c.AiImport())
	assert.NotNil(t, c.AiStepSort())
	assert.NotNil(t, c.IngredientParser())
	assert.NotNil(t, c.FdcSearch())
	assert.NotNil(t, c.ImportOpenData())
	assert.NotNil(t, c.ShareLinks())
	assert.NotNil(t, c.RecipeImports())
	assert.NotNil(t, c.ImportLogs())
	assert.NotNil(t, c.ExportLogs())
	assert.NotNil(t, c.BookmarkletImports())
	assert.NotNil(t, c.Syncs())
	assert.NotNil(t, c.SyncLogs())
	assert.NotNil(t, c.AppExport())
	assert.NotNil(t, c.CookLogs())
	assert.NotNil(t, c.ViewLogs())
	assert.NotNil(t, c.UserFiles())
	assert.NotNil(t, c.Automations())
	assert.NotNil(t, c.CustomFilters())
	assert.NotNil(t, c.ConnectorConfigs())
	assert.NotNil(t, c.SearchFields())
	assert.NotNil(t, c.SearchPreference())
	assert.NotNil(t, c.AiProviders())
	assert.NotNil(t, c.AiLogs())
	assert.NotNil(t, c.Localization())
	assert.NotNil(t, c.ServerSettings())
}

func TestClient_BaseURLOrigin(t *testing.T) {
	c, err := NewClient("http://localhost:8000")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8000", c.BaseURLOrigin())
}

func TestClient_DoJSON(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte(`{"id":1}`))
	}))
	t.Cleanup(server.Close)

	c, err := NewClient(server.URL, WithAccessToken("tok"))
	require.NoError(t, err)

	var out map[string]int
	require.NoError(t, c.DoJSON(context.Background(), "GET", "api/food/", nil, &out))
	assert.Equal(t, "Bearer tok", gotAuth)
	assert.Equal(t, 1, out["id"])
}

func TestClient_DoJSON_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"detail":"Not found."}`))
	}))
	t.Cleanup(server.Close)

	c, err := NewClient(server.URL)
	require.NoError(t, err)

	err = c.DoJSON(context.Background(), "GET", "api/food/99/", nil, nil)
	require.Error(t, err)
	var te *TandoorError
	require.ErrorAs(t, err, &te)
	assert.True(t, te.IsNotFound())
}

func TestClient_DoRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("BEGIN:VCALENDAR"))
	}))
	t.Cleanup(server.Close)

	c, err := NewClient(server.URL)
	require.NoError(t, err)

	resp, err := c.DoRaw(context.Background(), "GET", "api/meal-plan/ical/", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_DoRawNoAuth(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	c, err := NewClient(server.URL, WithAccessToken("tok"))
	require.NoError(t, err)

	resp, err := c.DoRawNoAuth(context.Background(), "POST", "api-token-auth/", strings.NewReader("username=x&password=y"))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Empty(t, gotAuth)
}

func TestClient_Auth(t *testing.T) {
	c, err := NewClient("http://example.test", WithAccessToken("tok"))
	require.NoError(t, err)
	assert.NotNil(t, c.Auth())
}
