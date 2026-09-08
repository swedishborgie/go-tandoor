// Package recipe provides typed list options.
package recipe

import (
	"net/url"
	"strconv"

	"github.com/swedishborgie/go-tandoor/pagination"
)

// ListOptions provides typed filtering for recipe list endpoints.
// It embeds pagination.ListOptions for pagination and ordering.
type ListOptions struct {
	pagination.ListOptions
	KeywordIDs   []int
	SpaceID      int
	RecipeBookID int
	UserID       int
	IsFavorite   *bool
}

// Values builds url.Values from the typed options, merging base pagination
// options and typed filters.
func (o ListOptions) Values() url.Values {
	v := o.ListOptions.Values()
	if len(o.KeywordIDs) > 0 {
		v.Set("keyword", joinInts(o.KeywordIDs))
	}
	if o.SpaceID != 0 {
		v.Set("space", strconv.Itoa(o.SpaceID))
	}
	if o.RecipeBookID != 0 {
		v.Set("recipe_book", strconv.Itoa(o.RecipeBookID))
	}
	if o.UserID != 0 {
		v.Set("user", strconv.Itoa(o.UserID))
	}
	if o.IsFavorite != nil {
		v.Set("is_favorite", strconv.FormatBool(*o.IsFavorite))
	}
	return v
}

// QueryString returns the query string for the options.
func (o ListOptions) QueryString() string {
	v := o.Values()
	s := v.Encode()
	// net/url encodes spaces as +, pagination expects same.
	return s
}

// ToPaginationOptions converts the typed options to a *pagination.ListOptions
// for backward compatibility. Typed fields are encoded into Extra.
func (o ListOptions) ToPaginationOptions() *pagination.ListOptions {
	base := o.ListOptions
	extra := make(map[string]string)
	if len(o.KeywordIDs) > 0 {
		extra["keyword"] = joinInts(o.KeywordIDs)
	}
	if o.SpaceID != 0 {
		extra["space"] = strconv.Itoa(o.SpaceID)
	}
	if o.RecipeBookID != 0 {
		extra["recipe_book"] = strconv.Itoa(o.RecipeBookID)
	}
	if o.UserID != 0 {
		extra["user"] = strconv.Itoa(o.UserID)
	}
	if o.IsFavorite != nil {
		extra["is_favorite"] = strconv.FormatBool(*o.IsFavorite)
	}
	if base.Extra == nil {
		base.Extra = extra
	} else {
		for k, v := range extra {
			base.Extra[k] = v
		}
	}
	return &base
}

func joinInts(ids []int) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.Itoa(id)
	}
	return joinStrings(s)
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for _, v := range ss[1:] {
		result += "," + v
	}
	return result
}
