package misc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

// CookLogService tests

func TestCookLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/cook-log/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"recipe":1,"servings":4,"rating":5}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, 4, *page.Results[0].Servings)
}

func TestCookLogService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/cook-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"recipe":1,"servings":4,"rating":5,"comment":"Delicious!"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Delicious!", *log.Comment)
}

func TestCookLogService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"recipe":1,"servings":2}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, *log.Servings)
}

func TestCookLogService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"servings":6}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewService(mock).Update(context.Background(), &CookLog{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 6, *log.Servings)
}

func TestCookLogService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewService(mock).Patch(context.Background(), &CookLog{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, log.ID)
}

func TestCookLogService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/cook-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// ViewLogService tests

func TestViewLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/view-log/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"recipe":2}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewViewLogService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
}

func TestViewLogService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/view-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"recipe":2,"created_at":"2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewViewLogService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-01T00:00:00Z", log.CreatedAt)
}

func TestViewLogService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"recipe":3}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewViewLogService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 3, *log.Recipe)
}

func TestViewLogService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/view-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewViewLogService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// UserFileService tests

func TestUserFileService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user-file/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"photo.jpg","file_size_kb":1024}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewUserFileService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "photo.jpg", page.Results[0].Name)
}

func TestUserFileService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user-file/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"photo.jpg","file_size_kb":1024,"file_download":"https://example.com/download/1/","preview":""}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	file, err := NewUserFileService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1024, file.FileSizeKB)
}

func TestUserFileService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"new.jpg","file_size_kb":512}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	file, err := NewUserFileService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "new.jpg", file.Name)
}

func TestUserFileService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/user-file/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewUserFileService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// AutomationService tests

func TestAutomationService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/automation/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"type":"FOOD_ALIAS","name":"Flour alias","disabled":false}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewAutomationService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "FOOD_ALIAS", page.Results[0].Type)
}

func TestAutomationService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/automation/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"type":"FOOD_ALIAS","name":"Flour alias","param_1":"flour","param_2":"wheat flour"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	auto, err := NewAutomationService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "flour", auto.Param1)
}

func TestAutomationService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"type":"UNIT_ALIAS","name":"Cups alias"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	auto, err := NewAutomationService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "UNIT_ALIAS", auto.Type)
}

func TestAutomationService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"disabled":true}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	auto, err := NewAutomationService(mock).Update(context.Background(), &Automation{ID: 1})
	require.NoError(t, err)
	assert.True(t, auto.Disabled)
}

func TestAutomationService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	auto, err := NewAutomationService(mock).Patch(context.Background(), &Automation{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, auto.ID)
}

func TestAutomationService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/automation/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewAutomationService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// CustomFilterService tests

func TestCustomFilterService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/custom-filter/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Favorites","search":"keywords:chicken OR keywords:pasta"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewCustomFilterService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "Favorites", page.Results[0].Name)
}

func TestCustomFilterService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/custom-filter/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Quick Meals","search":"working_time__lt:30"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cf, err := NewCustomFilterService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Quick Meals", cf.Name)
}

func TestCustomFilterService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"New Filter","search":"keywords:vegan"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cf, err := NewCustomFilterService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "New Filter", cf.Name)
}

func TestCustomFilterService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated Filter"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cf, err := NewCustomFilterService(mock).Update(context.Background(), &CustomFilter{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Updated Filter", cf.Name)
}

func TestCustomFilterService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cf, err := NewCustomFilterService(mock).Patch(context.Background(), &CustomFilter{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, cf.ID)
}

func TestCustomFilterService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/custom-filter/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewCustomFilterService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// ConnectorConfigService tests

func TestConnectorConfigService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/connector-config/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"HA","type":"HomeAssistant","enabled":true}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewConnectorConfigService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "HomeAssistant", page.Results[0].Type)
}

func TestConnectorConfigService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/connector-config/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"HA","type":"HomeAssistant","enabled":true,"todo_entity":"todo.shopping"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cc, err := NewConnectorConfigService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "todo.shopping", cc.TodoEntity)
}

func TestConnectorConfigService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"New Connector","enabled":true}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	cc, err := NewConnectorConfigService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "New Connector", cc.Name)
}

func TestConnectorConfigService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/connector-config/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewConnectorConfigService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// SearchFieldsService tests

func TestSearchFieldsService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/search-fields/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"name":"Name","field":"name"},{"id":2,"name":"Description","field":"description"}]`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	fields, err := NewSearchFieldsService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, fields, 2)
	assert.Equal(t, "Name", fields[0].Name)
}

func TestSearchFieldsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/search-fields/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Name","field":"name"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sf, err := NewSearchFieldsService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "name", sf.Field)
}

// SearchPreferenceService tests

func TestSearchPreferenceService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/search-preference/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"search":"plain","lookup":false,"trigram_threshold":0.2}]`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	prefs, err := NewSearchPreferenceService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, prefs, 1)
	assert.Equal(t, "plain", prefs[0].Search)
}

func TestSearchPreferenceService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/search-preference/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user":{"id":1},"search":"websearch","trigram_threshold":0.3}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sp, err := NewSearchPreferenceService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "websearch", sp.Search)
}

// AiProviderService tests

func TestAiProviderService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/ai-provider/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"OpenAI","model_name":"gpt-4","log_credit_cost":true}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewAiProviderService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "gpt-4", page.Results[0].ModelName)
}

func TestAiProviderService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/ai-provider/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"OpenAI","model_name":"gpt-4","url":"https://api.openai.com"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	ap, err := NewAiProviderService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "OpenAI", ap.Name)
}

func TestAiProviderService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"Anthropic","model_name":"claude-3","log_credit_cost":true}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	ap, err := NewAiProviderService(mock).Create(context.Background(), &AiProviderCreate{
		Name:      "Anthropic",
		APIKey:    "sk-test",
		ModelName: "claude-3",
	})
	require.NoError(t, err)
	assert.Equal(t, "claude-3", ap.ModelName)
}

func TestAiProviderService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated Provider"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	ap, err := NewAiProviderService(mock).Update(context.Background(), &AiProvider{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Updated Provider", ap.Name)
}

func TestAiProviderService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	ap, err := NewAiProviderService(mock).Patch(context.Background(), &AiProvider{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, ap.ID)
}

func TestAiProviderService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/ai-provider/1/", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewAiProviderService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// AiLogService tests

func TestAiLogService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/ai-log/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"function":"FILE_IMPORT","credit_cost":0.01,"input_tokens":100,"output_tokens":50}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewAiLogService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "FILE_IMPORT", page.Results[0].Function)
}

func TestAiLogService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/ai-log/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"function":"STEP_SORT","credit_cost":0.005}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	log, err := NewAiLogService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "STEP_SORT", log.Function)
}

// LocalizationService tests

func TestLocalizationService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/localization/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"code":"en","language":"English"},{"code":"de","language":"Deutsch"}]`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	result, err := NewLocalizationService(mock).List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "en", result[0].Code)
	assert.Equal(t, "Deutsch", result[1].Language)
}

// ServerSettingsService tests

func TestServerSettingsService_Settings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/server-settings/settings/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"shopping_min_autosync_interval":"0","disable_external_connectors":false,"hosted":false,"version":"1.5.10"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	settings, err := NewServerSettingsService(mock).Settings(context.Background())
	require.NoError(t, err)
	assert.False(t, settings.Hosted)
	assert.Equal(t, "1.5.10", settings.Version)
}
