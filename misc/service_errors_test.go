package misc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/space"
)

func mockStatus(t *testing.T, status int, body string) *testutil.MockExecutor {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return testutil.NewMockExecutor(server.URL)
}

func notFound(t *testing.T) *testutil.MockExecutor {
	return mockStatus(t, http.StatusNotFound, `{"detail":"Not found."}`)
}

func TestViewLogService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"id":1,"recipe":5}`))
	}))
	t.Cleanup(server.Close)

	l, err := NewViewLogService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &ViewLog{ID: 1})
	require.NoError(t, err)
	require.NotNil(t, l.Recipe)
	assert.Equal(t, 5, *l.Recipe)
}

func TestViewLogService_Update_Error(t *testing.T) {
	_, err := NewViewLogService(notFound(t)).Update(context.Background(), &ViewLog{ID: 99})
	require.Error(t, err)
}

func TestViewLogService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":1,"recipe":6}`))
	}))
	t.Cleanup(server.Close)

	l, err := NewViewLogService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &ViewLog{ID: 1})
	require.NoError(t, err)
	require.NotNil(t, l.Recipe)
	assert.Equal(t, 6, *l.Recipe)
}

func TestViewLogService_Patch_Error(t *testing.T) {
	_, err := NewViewLogService(notFound(t)).Patch(context.Background(), &ViewLog{ID: 99})
	require.Error(t, err)
}

func TestUserFileService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"id":1,"name":"renamed.png"}`))
	}))
	t.Cleanup(server.Close)

	f, err := NewUserFileService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &UserFile{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "renamed.png", f.Name)
}

func TestUserFileService_Update_Error(t *testing.T) {
	_, err := NewUserFileService(notFound(t)).Update(context.Background(), &UserFile{ID: 99})
	require.Error(t, err)
}

func TestUserFileService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":1,"name":"patched.png"}`))
	}))
	t.Cleanup(server.Close)

	f, err := NewUserFileService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &UserFile{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "patched.png", f.Name)
}

func TestUserFileService_Patch_Error(t *testing.T) {
	_, err := NewUserFileService(notFound(t)).Patch(context.Background(), &UserFile{ID: 99})
	require.Error(t, err)
}

func TestConnectorConfigService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"id":1,"name":"cc"}`))
	}))
	t.Cleanup(server.Close)

	cc, err := NewConnectorConfigService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &ConnectorConfig{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "cc", cc.Name)
}

func TestConnectorConfigService_Update_Error(t *testing.T) {
	_, err := NewConnectorConfigService(notFound(t)).Update(context.Background(), &ConnectorConfig{ID: 99})
	require.Error(t, err)
}

func TestConnectorConfigService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":1,"name":"cc2"}`))
	}))
	t.Cleanup(server.Close)

	cc, err := NewConnectorConfigService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &ConnectorConfig{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "cc2", cc.Name)
}

func TestConnectorConfigService_Patch_Error(t *testing.T) {
	_, err := NewConnectorConfigService(notFound(t)).Patch(context.Background(), &ConnectorConfig{ID: 99})
	require.Error(t, err)
}

func TestSearchPreferenceService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"search":"pasta","lookup":true}`))
	}))
	t.Cleanup(server.Close)

	sp, err := NewSearchPreferenceService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &SearchPreference{User: &space.User{ID: 1}})
	require.NoError(t, err)
	assert.True(t, sp.Lookup)
}

func TestSearchPreferenceService_Patch_Error(t *testing.T) {
	_, err := NewSearchPreferenceService(notFound(t)).Patch(context.Background(), &SearchPreference{User: &space.User{ID: 1}})
	require.Error(t, err)
}

func TestCookLogListOptions_Values(t *testing.T) {
	v := CookLogListOptions{ListOptions: pagination.ListOptions{Page: 2}}.Values()
	assert.Equal(t, "2", v.Get("page"))
}

func TestCookLogListOptions_ToPaginationOptions(t *testing.T) {
	p := CookLogListOptions{ListOptions: pagination.ListOptions{PageSize: 40}}.ToPaginationOptions()
	assert.Equal(t, 40, p.PageSize)
}

// TestMiscServicesErrorBranches drives every misc service method against a
// 404 server to cover the error branches.
func TestMiscServicesErrorBranches(t *testing.T) {
	e := notFound(t)
	ctx := context.Background()

	// CookLog (root Service).
	if _, err := NewService(e).List(ctx, &CookLogListOptions{}); err == nil {
		t.Error("CookLog List: expected error")
	}
	if _, err := NewService(e).Get(ctx, 1); err == nil {
		t.Error("CookLog Get: expected error")
	}
	if _, err := NewService(e).Create(ctx, &CookLog{}); err == nil {
		t.Error("CookLog Create: expected error")
	}
	if _, err := NewService(e).Update(ctx, &CookLog{ID: 1}); err == nil {
		t.Error("CookLog Update: expected error")
	}
	if _, err := NewService(e).Patch(ctx, &CookLog{ID: 1}); err == nil {
		t.Error("CookLog Patch: expected error")
	}
	if err := NewService(e).Delete(ctx, 1); err == nil {
		t.Error("CookLog Delete: expected error")
	}

	// ViewLog.
	if _, err := NewViewLogService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("ViewLog List: expected error")
	}
	if _, err := NewViewLogService(e).Get(ctx, 1); err == nil {
		t.Error("ViewLog Get: expected error")
	}
	if _, err := NewViewLogService(e).Create(ctx, &ViewLog{}); err == nil {
		t.Error("ViewLog Create: expected error")
	}
	if _, err := NewViewLogService(e).Update(ctx, &ViewLog{ID: 1}); err == nil {
		t.Error("ViewLog Update: expected error")
	}
	if _, err := NewViewLogService(e).Patch(ctx, &ViewLog{ID: 1}); err == nil {
		t.Error("ViewLog Patch: expected error")
	}
	if err := NewViewLogService(e).Delete(ctx, 1); err == nil {
		t.Error("ViewLog Delete: expected error")
	}

	// UserFile.
	if _, err := NewUserFileService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("UserFile List: expected error")
	}
	if _, err := NewUserFileService(e).Get(ctx, 1); err == nil {
		t.Error("UserFile Get: expected error")
	}
	if _, err := NewUserFileService(e).Create(ctx, &UserFile{}); err == nil {
		t.Error("UserFile Create: expected error")
	}
	if _, err := NewUserFileService(e).Update(ctx, &UserFile{ID: 1}); err == nil {
		t.Error("UserFile Update: expected error")
	}
	if _, err := NewUserFileService(e).Patch(ctx, &UserFile{ID: 1}); err == nil {
		t.Error("UserFile Patch: expected error")
	}
	if err := NewUserFileService(e).Delete(ctx, 1); err == nil {
		t.Error("UserFile Delete: expected error")
	}

	// Automation.
	if _, err := NewAutomationService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("Automation List: expected error")
	}
	if _, err := NewAutomationService(e).Get(ctx, 1); err == nil {
		t.Error("Automation Get: expected error")
	}
	if _, err := NewAutomationService(e).Create(ctx, &Automation{}); err == nil {
		t.Error("Automation Create: expected error")
	}
	if _, err := NewAutomationService(e).Update(ctx, &Automation{ID: 1}); err == nil {
		t.Error("Automation Update: expected error")
	}
	if _, err := NewAutomationService(e).Patch(ctx, &Automation{ID: 1}); err == nil {
		t.Error("Automation Patch: expected error")
	}
	if err := NewAutomationService(e).Delete(ctx, 1); err == nil {
		t.Error("Automation Delete: expected error")
	}

	// CustomFilter.
	if _, err := NewCustomFilterService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("CustomFilter List: expected error")
	}
	if _, err := NewCustomFilterService(e).Get(ctx, 1); err == nil {
		t.Error("CustomFilter Get: expected error")
	}
	if _, err := NewCustomFilterService(e).Create(ctx, &CustomFilter{}); err == nil {
		t.Error("CustomFilter Create: expected error")
	}
	if _, err := NewCustomFilterService(e).Update(ctx, &CustomFilter{ID: 1}); err == nil {
		t.Error("CustomFilter Update: expected error")
	}
	if _, err := NewCustomFilterService(e).Patch(ctx, &CustomFilter{ID: 1}); err == nil {
		t.Error("CustomFilter Patch: expected error")
	}
	if err := NewCustomFilterService(e).Delete(ctx, 1); err == nil {
		t.Error("CustomFilter Delete: expected error")
	}

	// ConnectorConfig.
	if _, err := NewConnectorConfigService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("ConnectorConfig List: expected error")
	}
	if _, err := NewConnectorConfigService(e).Get(ctx, 1); err == nil {
		t.Error("ConnectorConfig Get: expected error")
	}
	if _, err := NewConnectorConfigService(e).Create(ctx, &ConnectorConfig{}); err == nil {
		t.Error("ConnectorConfig Create: expected error")
	}
	if _, err := NewConnectorConfigService(e).Update(ctx, &ConnectorConfig{ID: 1}); err == nil {
		t.Error("ConnectorConfig Update: expected error")
	}
	if _, err := NewConnectorConfigService(e).Patch(ctx, &ConnectorConfig{ID: 1}); err == nil {
		t.Error("ConnectorConfig Patch: expected error")
	}
	if err := NewConnectorConfigService(e).Delete(ctx, 1); err == nil {
		t.Error("ConnectorConfig Delete: expected error")
	}

	// SearchFields.
	if _, err := NewSearchFieldsService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("SearchFields List: expected error")
	}
	if _, err := NewSearchFieldsService(e).Get(ctx, 1); err == nil {
		t.Error("SearchFields Get: expected error")
	}

	// SearchPreference.
	if _, err := NewSearchPreferenceService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("SearchPreference List: expected error")
	}
	if _, err := NewSearchPreferenceService(e).Get(ctx, 1); err == nil {
		t.Error("SearchPreference Get: expected error")
	}
	if _, err := NewSearchPreferenceService(e).Patch(ctx, &SearchPreference{User: &space.User{ID: 1}}); err == nil {
		t.Error("SearchPreference Patch: expected error")
	}

	// AiProvider.
	if _, err := NewAiProviderService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("AiProvider List: expected error")
	}
	if _, err := NewAiProviderService(e).Get(ctx, 1); err == nil {
		t.Error("AiProvider Get: expected error")
	}
	if _, err := NewAiProviderService(e).Create(ctx, &AiProviderCreate{}); err == nil {
		t.Error("AiProvider Create: expected error")
	}
	if _, err := NewAiProviderService(e).Update(ctx, &AiProvider{ID: 1}); err == nil {
		t.Error("AiProvider Update: expected error")
	}
	if _, err := NewAiProviderService(e).Patch(ctx, &AiProvider{ID: 1}); err == nil {
		t.Error("AiProvider Patch: expected error")
	}
	if err := NewAiProviderService(e).Delete(ctx, 1); err == nil {
		t.Error("AiProvider Delete: expected error")
	}

	// AiLog.
	if _, err := NewAiLogService(e).List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("AiLog List: expected error")
	}
	if _, err := NewAiLogService(e).Get(ctx, 1); err == nil {
		t.Error("AiLog Get: expected error")
	}

	// Localization + ServerSettings.
	if _, err := NewLocalizationService(e).List(ctx); err == nil {
		t.Error("Localization List: expected error")
	}
	if _, err := NewServerSettingsService(e).Settings(ctx); err == nil {
		t.Error("ServerSettings Settings: expected error")
	}
}
