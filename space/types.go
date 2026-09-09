// Package space contains types for Tandoor Organization resources.
//
// This covers Space (tenancy boundary), User, Group, Household,
// UserSpace (membership), and UserPreference.
package space

import (
	"time"
)

// Space represents a Tandoor tenancy boundary (multi-tenant isolation).
type Space struct {
	ID int `json:"id"`

	Name string `json:"name"`

	CreatedBy User      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`

	Message string `json:"message"`

	// Limits
	MaxRecipes       int `json:"max_recipes"`
	MaxFileStorageMB int `json:"max_file_storage_mb"`
	MaxUsers         int `json:"max_users"`

	AllowSharing bool `json:"allow_sharing"`
	Demo         bool `json:"demo"`

	// Counts (read-only computed fields)
	UserCount   int     `json:"user_count"`
	RecipeCount int     `json:"recipe_count"`
	FileSizeMB  float64 `json:"file_size_mb"`

	// Branding
	Image            *UserFile `json:"image"`
	NavLogo          *UserFile `json:"nav_logo"`
	SpaceTheme       string    `json:"space_theme"`
	CustomSpaceTheme *UserFile `json:"custom_space_theme"`
	NavBGColor       string    `json:"nav_bg_color"`
	NavTextColor     string    `json:"nav_text_color"`

	LogoColor32  *UserFile `json:"logo_color_32"`
	LogoColor128 *UserFile `json:"logo_color_128"`
	LogoColor144 *UserFile `json:"logo_color_144"`
	LogoColor180 *UserFile `json:"logo_color_180"`
	LogoColor192 *UserFile `json:"logo_color_192"`
	LogoColor512 *UserFile `json:"logo_color_512"`
	LogoColorSVG *UserFile `json:"logo_color_svg"`

	// AI settings
	AICreditsMonthly     int     `json:"ai_credits_monthly"`
	AICreditsBalance     float64 `json:"ai_credits_balance"`
	AIMonthlyCreditsUsed int     `json:"ai_monthly_credits_used"`
	AIEnabled            bool    `json:"ai_enabled"`

	SpaceSetupCompleted     bool `json:"space_setup_completed"`
	HouseholdSetupCompleted bool `json:"household_setup_completed"`
}

// UserFile is a minimal file reference used in branding fields.
type UserFile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// User represents a Tandoor user (read-only, no create/delete).
type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	IsStaff     bool   `json:"is_staff"`
	IsSuperuser bool   `json:"is_superuser"`
	IsActive    bool   `json:"is_active"`
}

// Group represents a Django auth group (read-only).
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Household represents a household within a space.
type Household struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserSpace represents a user's membership in a space.
// Cannot be created via the API; managed through invites.
type UserSpace struct {
	ID           int        `json:"id"`
	User         User       `json:"user"`
	Space        int        `json:"space"`
	Groups       []Group    `json:"groups"`
	Household    *Household `json:"household"`
	Active       bool       `json:"active"`
	InternalNote string     `json:"internal_note"`
	InviteLink   string     `json:"invite_link"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserSpaceBatchUpdate performs bulk operations on user spaces.
type UserSpaceBatchUpdate struct {
	UserSpaces []int `json:"user_spaces"`
	Household  *int  `json:"household,omitempty"`
	GroupSet   []int `json:"group_set,omitempty"`
}

// UserPreference represents per-user settings for a space.
type UserPreference struct {
	User                       User      `json:"user"`
	Image                      *UserFile `json:"image"`
	Theme                      string    `json:"theme"`
	NavBGColor                 string    `json:"nav_bg_color"`
	NavTextColor               string    `json:"nav_text_color"`
	NavShowLogo                bool      `json:"nav_show_logo"`
	DefaultUnit                *int      `json:"default_unit"`
	DefaultPage                int       `json:"default_page"`
	UseFractions               bool      `json:"use_fractions"`
	UseKJ                      bool      `json:"use_kj"`
	NavSticky                  bool      `json:"nav_sticky"`
	IngredientDecimals         int       `json:"ingredient_decimals"`
	Comments                   bool      `json:"comments"`
	ShoppingAutoSync           bool      `json:"shopping_auto_sync"`
	MealplanAutoaddShopping    bool      `json:"mealplan_autoadd_shopping"`
	ShoppingRecentDays         int       `json:"shopping_recent_days"`
	CsvDelim                   string    `json:"csv_delim"`
	CsvPrefix                  string    `json:"csv_prefix"`
	DefaultDelay               int       `json:"default_delay"`
	MealplanAutoincludeRelated bool      `json:"mealplan_autoinclude_related"`
	MealplanAutoexcludeOnhand  bool      `json:"mealplan_autoexclude_onhand"`
	FilterToSupermarket        bool      `json:"filter_to_supermarket"`
	ShoppingAddOnhand          bool      `json:"shopping_add_onhand"`
	LeftHanded                 bool      `json:"left_handed"`
	ShowStepIngredients        bool      `json:"show_step_ingredients"`
}

// InviteLink represents an invitation to join a space.
type InviteLink struct {
	ID           int        `json:"id"`
	UUID         string     `json:"uuid"`
	Email        string     `json:"email"`
	Group        *Group     `json:"group"`
	Household    *Household `json:"household"`
	ValidUntil   time.Time  `json:"valid_until"`
	UsedBy       *User      `json:"used_by"`
	Reusable     bool       `json:"reusable"`
	InternalNote string     `json:"internal_note"`
	CreatedBy    User       `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	EmailSent    bool       `json:"email_sent"`
}

// InviteLinkRequest is the request body for creating an invite link.
type InviteLinkRequest struct {
	Email string `json:"email"`
	// GroupID is the group invited users get (required; see group_list).
	GroupID      int    `json:"group_id"`
	HouseholdID  *int   `json:"household_id,omitempty"`
	Reusable     bool   `json:"reusable"`
	InternalNote string `json:"internal_note,omitempty"`
	// ValidUntil is the expiry date (YYYY-MM-DD). The API field is a plain
	// date, so it is kept as a string rather than time.Time (whose JSON
	// form is a datetime the serializer rejects). Nil uses the server
	// default.
	ValidUntil *string `json:"valid_until,omitempty"`
}
