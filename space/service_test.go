package space

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swedishborgie/go-tandoor/internal/testutil"
)

// SpaceService tests

func TestSpaceService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/space/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":2,"results":[{"id":1,"name":"Home"},{"id":2,"name":"Work"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Count)
	assert.Equal(t, "Home", page.Results[0].Name)
}

func TestSpaceService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/space/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Home","user_count":3}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sp, err := NewService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 3, sp.UserCount)
}

func TestSpaceService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":10,"name":"New Space"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sp, err := NewService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 10, sp.ID)
}

func TestSpaceService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Updated"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sp, err := NewService(mock).Update(context.Background(), &Space{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Updated", sp.Name)
}

func TestSpaceService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sp, err := NewService(mock).Patch(context.Background(), &Space{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, sp.ID)
}

func TestSpaceService_Current(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/space/current/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Home"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	sp, err := NewService(mock).Current(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Home", sp.Name)
}

// UserService tests

func TestUserService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"username":"admin"},{"id":2,"username":"guest"}]`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	users, err := NewUserService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "admin", users[0].Username)
}

func TestUserService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"username":"admin","is_staff":true}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	u, err := NewUserService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, u.IsStaff)
}

func TestUserService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	u, err := NewUserService(mock).Patch(context.Background(), &User{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, u.ID)
}

// GroupService tests

func TestGroupService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/group/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"name":"admin"},{"id":2,"name":"user"}]`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	groups, err := NewGroupService(mock).List(context.Background())
	require.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.Equal(t, "admin", groups[0].Name)
}

func TestGroupService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/group/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"admin"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	g, err := NewGroupService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "admin", g.Name)
}

// HouseholdService tests

func TestHouseholdService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/household/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"name":"Main"}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewHouseholdService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.Equal(t, "Main", page.Results[0].Name)
}

func TestHouseholdService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":5,"name":"Guest Room"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	h, err := NewHouseholdService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 5, h.ID)
}

func TestHouseholdService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewHouseholdService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// UserSpaceService tests

func TestUserSpaceService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user-space/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"user":{"id":1},"active":true}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewUserSpaceService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.True(t, page.Results[0].Active)
}

func TestUserSpaceService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"active":false}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	us, err := NewUserSpaceService(mock).Update(context.Background(), &UserSpace{ID: 1})
	require.NoError(t, err)
	assert.False(t, us.Active)
}

func TestUserSpaceService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewUserSpaceService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestUserSpaceService_AllPersonal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user-space/all_personal/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"space":1},{"id":2,"space":2}]`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	us, err := NewUserSpaceService(mock).AllPersonal(context.Background())
	require.NoError(t, err)
	assert.Len(t, us, 2)
	assert.Equal(t, 2, us[1].Space)
}

func TestUserSpaceService_BatchUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/user-space/batch_update/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewUserSpaceService(mock).BatchUpdate(context.Background(), &UserSpaceBatchUpdate{UserSpaces: []int{1, 2}})
	assert.NoError(t, err)
}

// UserPreferenceService tests

func TestUserPreferenceService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user-preference/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user":{"id":1},"theme":"dark","use_fractions":true}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	up, err := NewUserPreferenceService(mock).Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "dark", up.Theme)
	assert.True(t, up.UseFractions)
}

func TestUserPreferenceService_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user":{"id":1},"theme":"light"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	up, err := NewUserPreferenceService(mock).Patch(context.Background(), &UserPreference{})
	require.NoError(t, err)
	assert.Equal(t, "light", up.Theme)
}

// InviteLinkService tests

func TestInviteLinkService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/invite-link/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1,"results":[{"id":1,"email":"test@example.com","reusable":true}]}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	page, err := NewInviteLinkService(mock).List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Count)
	assert.True(t, page.Results[0].Reusable)
}

func TestInviteLinkService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/invite-link/1/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"uuid":"abc123","email":"test@example.com"}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	il, err := NewInviteLinkService(mock).Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "abc123", il.UUID)
}

func TestInviteLinkService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"uuid":"abc123","email_sent":true}`))
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	il, err := NewInviteLinkService(mock).Create(context.Background(), nil)
	require.NoError(t, err)
	assert.True(t, il.EmailSent)
}

func TestInviteLinkService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	mock := testutil.NewMockExecutor(server.URL)

	err := NewInviteLinkService(mock).Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func TestInviteListOptions_QueryString(t *testing.T) {
	trueVal := true
	opts := &InviteListOptions{
		Page:         2,
		PageSize:     10,
		InternalNote: "test",
		Used:         &trueVal,
	}
	qs := opts.QueryString()
	assert.Contains(t, qs, "page=2")
	assert.Contains(t, qs, "page_size=10")
	assert.Contains(t, qs, "internal_note=test")
	assert.Contains(t, qs, "used=true")
}

func TestInviteListOptions_Empty(t *testing.T) {
	assert.True(t, (&InviteListOptions{}).Empty())
	assert.True(t, (*InviteListOptions)(nil).Empty())
	falseVal := false
	opts := &InviteListOptions{Page: 1}
	assert.False(t, opts.Empty())
	opts2 := &InviteListOptions{Used: &falseVal}
	assert.False(t, opts2.Empty())
}
