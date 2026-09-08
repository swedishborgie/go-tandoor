package fdc

import (
	"context"
	"fmt"
	"net/url"
)

// SearchFoods searches for foods using keywords.
//
// query is the search string (supports Elasticsearch operators).
// dataType filters by data type (pass nil for all types).
// pageSize sets the page size (1-200, nil for default of 50).
// pageNumber sets the page number (0-indexed, nil for first page).
// sortBy sets the sort field (pass "" for default).
// sortOrder sets the sort direction (pass "" for default).
// brandOwner filters by brand owner (pass "" for no filter, applies to Branded foods only).
func (c *Client) SearchFoods(
	ctx context.Context,
	query string,
	dataType []DataType,
	pageSize *int,
	pageNumber *int,
	sortBy SortField,
	sortOrder SortOrder,
	brandOwner string,
) (*SearchResponse, error) {
	q := url.Values{}
	q.Set("query", query)
	for _, dt := range dataType {
		q.Add("dataType", string(dt))
	}
	if pageSize != nil {
		q.Add("pageSize", fmt.Sprintf("%d", *pageSize))
	}
	if pageNumber != nil {
		q.Add("pageNumber", fmt.Sprintf("%d", *pageNumber))
	}
	if sortBy != "" {
		q.Add("sortBy", string(sortBy))
	}
	if sortOrder != "" {
		q.Add("sortOrder", string(sortOrder))
	}
	if brandOwner != "" {
		q.Add("brandOwner", brandOwner)
	}

	var resp SearchResponse
	if err := c.do(ctx, "GET", "v1/foods/search?"+q.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PostFoodsSearch searches for foods using keywords via a JSON request body.
// This is the POST variant of SearchFoods.
func (c *Client) PostFoodsSearch(ctx context.Context, criteria *FoodSearchCriteria) (*SearchResponse, error) {
	if criteria == nil || criteria.Query == "" {
		return nil, fmt.Errorf("fdc: criteria with a non-empty query is required")
	}

	var resp SearchResponse
	if err := c.do(ctx, "POST", "v1/foods/search", criteria, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
