package space

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
	"github.com/swedishborgie/go-tandoor/pagination"
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

func TestHouseholdService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/household/2/", r.URL.Path)
		w.Write([]byte(`{"id":2,"name":"Borg"}`))
	}))
	t.Cleanup(server.Close)

	h, err := NewHouseholdService(testutil.NewMockExecutor(server.URL)).Get(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, "Borg", h.Name)
}

func TestHouseholdService_Get_Error(t *testing.T) {
	_, err := NewHouseholdService(notFound(t)).Get(context.Background(), 99)
	require.Error(t, err)
}

func TestHouseholdService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"id":2,"name":"Renamed"}`))
	}))
	t.Cleanup(server.Close)

	h, err := NewHouseholdService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &Household{ID: 2, Name: "Renamed"})
	require.NoError(t, err)
	assert.Equal(t, "Renamed", h.Name)
}

func TestHouseholdService_Update_Error(t *testing.T) {
	_, err := NewHouseholdService(notFound(t)).Update(context.Background(), &Household{ID: 99})
	require.Error(t, err)
}

func TestHouseholdService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":2,"name":"Patched"}`))
	}))
	t.Cleanup(server.Close)

	h, err := NewHouseholdService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &Household{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, "Patched", h.Name)
}

func TestHouseholdService_Patch_Error(t *testing.T) {
	_, err := NewHouseholdService(notFound(t)).Patch(context.Background(), &Household{ID: 99})
	require.Error(t, err)
}

func TestUserSpaceService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/user-space/3/", r.URL.Path)
		w.Write([]byte(`{"id":3,"space":1,"active":true}`))
	}))
	t.Cleanup(server.Close)

	us, err := NewUserSpaceService(testutil.NewMockExecutor(server.URL)).Get(context.Background(), 3)
	require.NoError(t, err)
	assert.True(t, us.Active)
}

func TestUserSpaceService_Get_Error(t *testing.T) {
	_, err := NewUserSpaceService(notFound(t)).Get(context.Background(), 99)
	require.Error(t, err)
}

func TestUserSpaceService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":3,"active":false}`))
	}))
	t.Cleanup(server.Close)

	us, err := NewUserSpaceService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &UserSpace{ID: 3})
	require.NoError(t, err)
	assert.False(t, us.Active)
}

func TestUserSpaceService_Patch_Error(t *testing.T) {
	_, err := NewUserSpaceService(notFound(t)).Patch(context.Background(), &UserSpace{ID: 99})
	require.Error(t, err)
}

func TestInviteLinkService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/invite-link/4/", r.URL.Path)
		w.Write([]byte(`{"id":4,"email":"a@b.c"}`))
	}))
	t.Cleanup(server.Close)

	il, err := NewInviteLinkService(testutil.NewMockExecutor(server.URL)).Update(context.Background(), &InviteLink{ID: 4})
	require.NoError(t, err)
	assert.Equal(t, "a@b.c", il.Email)
}

func TestInviteLinkService_Update_Error(t *testing.T) {
	_, err := NewInviteLinkService(notFound(t)).Update(context.Background(), &InviteLink{ID: 99})
	require.Error(t, err)
}

func TestInviteLinkService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"id":4,"email":"a@b.c"}`))
	}))
	t.Cleanup(server.Close)

	il, err := NewInviteLinkService(testutil.NewMockExecutor(server.URL)).Patch(context.Background(), &InviteLink{ID: 4})
	require.NoError(t, err)
	assert.Equal(t, "a@b.c", il.Email)
}

func TestInviteLinkService_Patch_Error(t *testing.T) {
	_, err := NewInviteLinkService(notFound(t)).Patch(context.Background(), &InviteLink{ID: 99})
	require.Error(t, err)
}

func TestSpaceListOptions_Values(t *testing.T) {
	v := ListOptions{ListOptions: pagination.ListOptions{Page: 2}}.Values()
	assert.Equal(t, "2", v.Get("page"))
}

func TestSpaceListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{ListOptions: pagination.ListOptions{PageSize: 75}}.ToPaginationOptions()
	assert.Equal(t, 75, p.PageSize)
}

// TestSpaceErrorBranches drives every space service method against a 404
// server to cover the error branches.
func TestSpaceErrorBranches(t *testing.T) {
	e := notFound(t)
	ctx := context.Background()

	s := NewService(e)
	if _, err := s.List(ctx, &ListOptions{}); err == nil {
		t.Error("Space List: expected error")
	}
	if _, err := s.Get(ctx, 1); err == nil {
		t.Error("Space Get: expected error")
	}
	if _, err := s.Create(ctx, &Space{}); err == nil {
		t.Error("Space Create: expected error")
	}
	if _, err := s.Update(ctx, &Space{ID: 1}); err == nil {
		t.Error("Space Update: expected error")
	}
	if _, err := s.Patch(ctx, &Space{ID: 1}); err == nil {
		t.Error("Space Patch: expected error")
	}
	if _, err := s.Current(ctx); err == nil {
		t.Error("Space Current: expected error")
	}

	u := NewUserService(e)
	if _, err := u.List(ctx, nil); err == nil {
		t.Error("User List: expected error")
	}
	if _, err := u.Get(ctx, 1); err == nil {
		t.Error("User Get: expected error")
	}
	if _, err := u.Patch(ctx, &User{ID: 1}); err == nil {
		t.Error("User Patch: expected error")
	}

	g := NewGroupService(e)
	if _, err := g.List(ctx); err == nil {
		t.Error("Group List: expected error")
	}
	if _, err := g.Get(ctx, 1); err == nil {
		t.Error("Group Get: expected error")
	}

	h := NewHouseholdService(e)
	if _, err := h.List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("Household List: expected error")
	}
	if err := h.Delete(ctx, 1); err == nil {
		t.Error("Household Delete: expected error")
	}

	us := NewUserSpaceService(e)
	if _, err := us.List(ctx, &pagination.ListOptions{}); err == nil {
		t.Error("UserSpace List: expected error")
	}
	if _, err := us.Update(ctx, &UserSpace{ID: 1}); err == nil {
		t.Error("UserSpace Update: expected error")
	}
	if err := us.Delete(ctx, 1); err == nil {
		t.Error("UserSpace Delete: expected error")
	}
	if _, err := us.AllPersonal(ctx); err == nil {
		t.Error("UserSpace AllPersonal: expected error")
	}
	if err := us.BatchUpdate(ctx, &UserSpaceBatchUpdate{UserSpaces: []int{1}}); err == nil {
		t.Error("UserSpace BatchUpdate: expected error")
	}

	up := NewUserPreferenceService(e)
	if _, err := up.Get(ctx); err == nil {
		t.Error("UserPreference Get: expected error")
	}
	if _, err := up.Patch(ctx, &UserPreference{}); err == nil {
		t.Error("UserPreference Patch: expected error")
	}

	il := NewInviteLinkService(e)
	if _, err := il.List(ctx, &InviteListOptions{}); err == nil {
		t.Error("InviteLink List: expected error")
	}
	if _, err := il.Get(ctx, 1); err == nil {
		t.Error("InviteLink Get: expected error")
	}
	if _, err := il.Create(ctx, &InviteLinkRequest{Email: "a@b.c", GroupID: 1}); err == nil {
		t.Error("InviteLink Create: expected error")
	}
	if err := il.Delete(ctx, 1); err == nil {
		t.Error("InviteLink Delete: expected error")
	}
}
