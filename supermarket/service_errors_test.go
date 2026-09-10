package supermarket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/swedishborgie/go-tandoor/internal/testutil"
	"github.com/swedishborgie/go-tandoor/pagination"
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

// TestSupermarketErrorBranches drives every supermarket service method
// against a 404 server to cover the error branches.
func TestSupermarketErrorBranches(t *testing.T) {
	e := errExecutor(t)
	ctx := context.Background()

	s := NewService(e)
	if _, err := s.List(ctx, &ListOptions{}); err == nil {
		t.Error("Supermarket List: expected error")
	}
	if _, err := s.Get(ctx, 1); err == nil {
		t.Error("Supermarket Get: expected error")
	}
	if _, err := s.Create(ctx, &Supermarket{}); err == nil {
		t.Error("Supermarket Create: expected error")
	}
	if _, err := s.Update(ctx, &Supermarket{ID: 1}); err == nil {
		t.Error("Supermarket Update: expected error")
	}
	if _, err := s.Patch(ctx, &Supermarket{ID: 1}); err == nil {
		t.Error("Supermarket Patch: expected error")
	}
	if err := s.Delete(ctx, 1); err == nil {
		t.Error("Supermarket Delete: expected error")
	}

	cat := NewCategoryService(e)
	if _, err := cat.List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("Category List: expected error")
	}
	if _, err := cat.Get(ctx, 1); err == nil {
		t.Error("Category Get: expected error")
	}
	if _, err := cat.Create(ctx, &Category{}); err == nil {
		t.Error("Category Create: expected error")
	}
	if _, err := cat.Update(ctx, &Category{ID: 1}); err == nil {
		t.Error("Category Update: expected error")
	}
	if _, err := cat.Patch(ctx, &Category{ID: 1}); err == nil {
		t.Error("Category Patch: expected error")
	}
	if err := cat.Delete(ctx, 1); err == nil {
		t.Error("Category Delete: expected error")
	}
	if _, err := cat.Merge(ctx, 1, 2); err == nil {
		t.Error("Category Merge: expected error")
	}

	rel := NewCategoryRelationService(e)
	if _, err := rel.List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("CategoryRelation List: expected error")
	}
	if _, err := rel.Get(ctx, 1); err == nil {
		t.Error("CategoryRelation Get: expected error")
	}
	if _, err := rel.Create(ctx, &CategoryRelation{}); err == nil {
		t.Error("CategoryRelation Create: expected error")
	}
	if _, err := rel.Update(ctx, &CategoryRelation{ID: 1}); err == nil {
		t.Error("CategoryRelation Update: expected error")
	}
	if _, err := rel.Patch(ctx, &CategoryRelation{ID: 1}); err == nil {
		t.Error("CategoryRelation Patch: expected error")
	}
	if err := rel.Delete(ctx, 1); err == nil {
		t.Error("CategoryRelation Delete: expected error")
	}
}
