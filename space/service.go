package space

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/swedishborgie/go-tandoor/internal/executor"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// Service provides access to Space API endpoints.
type Service struct {
	exec executor.Executor
}

// NewService creates a new SpaceService.
func NewService(e executor.Executor) *Service {
	return &Service{exec: e}
}

// List returns a paginated list of spaces the user has access to.
func (s *Service) List(ctx context.Context, opts *ListOptions) (*pagination.Paginated[Space], error) {
	path := "api/space/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Space]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single space by ID.
func (s *Service) Get(ctx context.Context, id int) (*Space, error) {
	var sp Space
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/space/%d/", id), nil, &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// Create creates a new space.
func (s *Service) Create(ctx context.Context, sp *Space) (*Space, error) {
	var result Space
	if err := s.exec.DoJSON(ctx, "POST", "api/space/", sp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a space.
func (s *Service) Update(ctx context.Context, sp *Space) (*Space, error) {
	var result Space
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/space/%d/", sp.ID), sp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a space.
func (s *Service) Patch(ctx context.Context, sp *Space) (*Space, error) {
	var result Space
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/space/%d/", sp.ID), sp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Current returns the currently active space.
func (s *Service) Current(ctx context.Context) (*Space, error) {
	var sp Space
	if err := s.exec.DoJSON(ctx, "GET", "api/space/current/", nil, &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// UserService provides access to User API endpoints.
//
// User endpoints are read-only (GET, PATCH only). No create/delete.
type UserService struct {
	exec executor.Executor
}

// NewUserService creates a new UserService.
func NewUserService(e executor.Executor) *UserService {
	return &UserService{exec: e}
}

// List returns a flat list of users in the current space.
// Pagination is disabled for this endpoint.
func (s *UserService) List(ctx context.Context, filterList []int) ([]User, error) {
	v := url.Values{}
	for _, id := range filterList {
		v.Add("filter_list", fmt.Sprint(id))
	}
	path := "api/user/"
	if qs := v.Encode(); qs != "" {
		path += "?" + qs
	}
	var users []User
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// Get retrieves a single user by ID.
func (s *UserService) Get(ctx context.Context, id int) (*User, error) {
	var u User
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/user/%d/", id), nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Patch performs a partial update on a user.
func (s *UserService) Patch(ctx context.Context, u *User) (*User, error) {
	var result User
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/user/%d/", u.ID), u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GroupService provides access to Group API endpoints.
//
// Group endpoints are read-only (GET only).
type GroupService struct {
	exec executor.Executor
}

// NewGroupService creates a new GroupService.
func NewGroupService(e executor.Executor) *GroupService {
	return &GroupService{exec: e}
}

// List returns a flat list of groups.
// Pagination is disabled for this endpoint.
func (s *GroupService) List(ctx context.Context) ([]Group, error) {
	var groups []Group
	if err := s.exec.DoJSON(ctx, "GET", "api/group/", nil, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// Get retrieves a single group by ID.
func (s *GroupService) Get(ctx context.Context, id int) (*Group, error) {
	var g Group
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/group/%d/", id), nil, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// HouseholdService provides access to Household API endpoints.
type HouseholdService struct {
	exec executor.Executor
}

// NewHouseholdService creates a new HouseholdService.
func NewHouseholdService(e executor.Executor) *HouseholdService {
	return &HouseholdService{exec: e}
}

// List returns a paginated list of households in the current space.
func (s *HouseholdService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[Household], error) {
	path := "api/household/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[Household]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single household by ID.
func (s *HouseholdService) Get(ctx context.Context, id int) (*Household, error) {
	var h Household
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/household/%d/", id), nil, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

// Create creates a new household.
func (s *HouseholdService) Create(ctx context.Context, h *Household) (*Household, error) {
	var result Household
	if err := s.exec.DoJSON(ctx, "POST", "api/household/", h, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a household.
func (s *HouseholdService) Update(ctx context.Context, h *Household) (*Household, error) {
	var result Household
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/household/%d/", h.ID), h, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a household.
func (s *HouseholdService) Patch(ctx context.Context, h *Household) (*Household, error) {
	var result Household
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/household/%d/", h.ID), h, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a household.
func (s *HouseholdService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/household/%d/", id), nil, nil)
}

// UserSpaceService provides access to UserSpace (membership) API endpoints.
//
// UserSpace cannot be created via the API; membership is managed through invites.
type UserSpaceService struct {
	exec executor.Executor
}

// NewUserSpaceService creates a new UserSpaceService.
func NewUserSpaceService(e executor.Executor) *UserSpaceService {
	return &UserSpaceService{exec: e}
}

// List returns a paginated list of user space memberships.
// Admins see all users; regular users see only themselves.
func (s *UserSpaceService) List(ctx context.Context, opts *pagination.ListOptions) (*pagination.Paginated[UserSpace], error) {
	path := "api/user-space/"
	if opts != nil && opts.QueryString() != "" {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[UserSpace]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single user space membership by ID.
func (s *UserSpaceService) Get(ctx context.Context, id int) (*UserSpace, error) {
	var us UserSpace
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/user-space/%d/", id), nil, &us); err != nil {
		return nil, err
	}
	return &us, nil
}

// Update replaces a user space membership.
func (s *UserSpaceService) Update(ctx context.Context, us *UserSpace) (*UserSpace, error) {
	var result UserSpace
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/user-space/%d/", us.ID), us, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on a user space membership.
func (s *UserSpaceService) Patch(ctx context.Context, us *UserSpace) (*UserSpace, error) {
	var result UserSpace
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/user-space/%d/", us.ID), us, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a user from a space.
func (s *UserSpaceService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/user-space/%d/", id), nil, nil)
}

// AllPersonal returns all user space memberships for the current user
// across all spaces they belong to.
func (s *UserSpaceService) AllPersonal(ctx context.Context) ([]UserSpace, error) {
	var us []UserSpace
	if err := s.exec.DoJSON(ctx, "GET", "api/user-space/all_personal/", nil, &us); err != nil {
		return nil, err
	}
	return us, nil
}

// BatchUpdate performs bulk operations on user spaces.
func (s *UserSpaceService) BatchUpdate(ctx context.Context, bu *UserSpaceBatchUpdate) error {
	return s.exec.DoJSON(ctx, "PUT", "api/user-space/batch_update/", bu, nil)
}

// UserPreferenceService provides access to UserPreference API endpoints.
//
// UserPreference cannot be created via the API.
type UserPreferenceService struct {
	exec executor.Executor
}

// NewUserPreferenceService creates a new UserPreferenceService.
func NewUserPreferenceService(e executor.Executor) *UserPreferenceService {
	return &UserPreferenceService{exec: e}
}

// Get returns the current user's preferences for this space.
func (s *UserPreferenceService) Get(ctx context.Context) (*UserPreference, error) {
	var up UserPreference
	if err := s.exec.DoJSON(ctx, "GET", "api/user-preference/", nil, &up); err != nil {
		return nil, err
	}
	return &up, nil
}

// Patch performs a partial update on the user's preferences.
func (s *UserPreferenceService) Patch(ctx context.Context, up *UserPreference) (*UserPreference, error) {
	var result UserPreference
	if err := s.exec.DoJSON(ctx, "PATCH", "api/user-preference/", up, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// InviteLinkService provides access to Invite Link API endpoints.
type InviteLinkService struct {
	exec executor.Executor
}

// NewInviteLinkService creates a new InviteLinkService.
func NewInviteLinkService(e executor.Executor) *InviteLinkService {
	return &InviteLinkService{exec: e}
}

// List returns a paginated list of invite links.
// By default returns only unused invites. Use Unused filter to control this.
func (s *InviteLinkService) List(ctx context.Context, opts *InviteListOptions) (*pagination.Paginated[InviteLink], error) {
	path := "api/invite-link/"
	if opts != nil && !opts.Empty() {
		path += "?" + opts.QueryString()
	}
	var page pagination.Paginated[InviteLink]
	if err := s.exec.DoJSON(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get retrieves a single invite link by ID.
func (s *InviteLinkService) Get(ctx context.Context, id int) (*InviteLink, error) {
	var il InviteLink
	if err := s.exec.DoJSON(ctx, "GET", fmt.Sprintf("api/invite-link/%d/", id), nil, &il); err != nil {
		return nil, err
	}
	return &il, nil
}

// Create creates a new invite link.
func (s *InviteLinkService) Create(ctx context.Context, req *InviteLinkRequest) (*InviteLink, error) {
	var result InviteLink
	if err := s.exec.DoJSON(ctx, "POST", "api/invite-link/", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces an invite link.
func (s *InviteLinkService) Update(ctx context.Context, il *InviteLink) (*InviteLink, error) {
	var result InviteLink
	if err := s.exec.DoJSON(ctx, "PUT", fmt.Sprintf("api/invite-link/%d/", il.ID), il, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Patch performs a partial update on an invite link.
func (s *InviteLinkService) Patch(ctx context.Context, il *InviteLink) (*InviteLink, error) {
	var result InviteLink
	if err := s.exec.DoJSON(ctx, "PATCH", fmt.Sprintf("api/invite-link/%d/", il.ID), il, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes an invite link.
func (s *InviteLinkService) Delete(ctx context.Context, id int) error {
	return s.exec.DoJSON(ctx, "DELETE", fmt.Sprintf("api/invite-link/%d/", id), nil, nil)
}

// InviteListOptions provides filtering options for InviteLink.List().
type InviteListOptions struct {
	Page         int
	PageSize     int
	InternalNote string
	Used         *bool // nil = unused only (default), true = all, false = unused only
}

// Empty returns true if no options are set.
func (o *InviteListOptions) Empty() bool {
	if o == nil {
		return true
	}
	return o.Page == 0 && o.PageSize == 0 && o.InternalNote == "" && o.Used == nil
}

// QueryString returns the encoded query string for List options.
func (o *InviteListOptions) QueryString() string {
	if o == nil || o.Empty() {
		return ""
	}
	v := url.Values{}
	if o.Page > 0 {
		v.Set("page", fmt.Sprint(o.Page))
	}
	if o.PageSize > 0 {
		v.Set("page_size", fmt.Sprint(o.PageSize))
	}
	if o.InternalNote != "" {
		v.Set("internal_note", o.InternalNote)
	}
	if o.Used != nil {
		v.Set("used", fmt.Sprint(*o.Used))
	}
	s := v.Encode()
	s = strings.TrimPrefix(s, "?")
	return s
}
