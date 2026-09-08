package fdc

import (
	"fmt"
	"net/url"

	"context"
)

// ListFoods returns a paged list of foods in abridged format.
//
// dataType filters by data type (pass nil for all types).
// pageSize sets the page size (1-200, nil for default of 50).
// pageNumber sets the page number (0-indexed, nil for first page).
// sortBy sets the sort field (pass "" for default).
// sortOrder sets the sort direction (pass "" for default).
func (c *Client) ListFoods(
	ctx context.Context,
	dataType []DataType,
	pageSize *int,
	pageNumber *int,
	sortBy SortField,
	sortOrder SortOrder,
) ([]AbridgedFood, error) {
	q := url.Values{}
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

	var foods []AbridgedFood
	if err := c.do(ctx, "GET", "v1/foods/list?"+q.Encode(), nil, &foods); err != nil {
		return nil, err
	}
	return foods, nil
}

// PostFoodsList returns a paged list of foods in abridged format using a JSON
// request body. This is the POST variant of ListFoods.
func (c *Client) PostFoodsList(ctx context.Context, criteria *FoodListCriteria) ([]AbridgedFood, error) {
	if criteria == nil {
		return nil, fmt.Errorf("fdc: criteria is required")
	}

	var foods []AbridgedFood
	if err := c.do(ctx, "POST", "v1/foods/list", criteria, &foods); err != nil {
		return nil, err
	}
	return foods, nil
}
