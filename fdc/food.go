package fdc

import (
	"context"
	"fmt"
	"net/url"
)

// GetFood retrieves a single food item by FDC ID.
//
// format controls the level of detail (abridged or full). Pass "" for default (full).
// nutrients is a list of nutrient numbers to filter by (up to 25). Pass nil for all.
func (c *Client) GetFood(ctx context.Context, fdcID int, format Format, nutrients []int) (*Food, error) {
	q := url.Values{}
	if format != "" {
		q.Set("format", string(format))
	}
	for _, n := range nutrients {
		q.Add("nutrients", fmt.Sprintf("%d", n))
	}

	queryStr := ""
	if len(q) > 0 {
		queryStr = "?" + q.Encode()
	}

	var food Food
	if err := c.do(ctx, "GET", fmt.Sprintf("v1/food/%d%s", fdcID, queryStr), nil, &food); err != nil {
		return nil, err
	}
	return &food, nil
}

// GetFoods retrieves multiple food items by FDC IDs (up to 20).
//
// format controls the level of detail (abridged or full). Pass "" for default (full).
// nutrients is a list of nutrient numbers to filter by (up to 25). Pass nil for all.
func (c *Client) GetFoods(ctx context.Context, fdcIDs []int, format Format, nutrients []int) ([]Food, error) {
	q := url.Values{}
	for _, id := range fdcIDs {
		q.Add("fdcIds", fmt.Sprintf("%d", id))
	}
	if format != "" {
		q.Set("format", string(format))
	}
	for _, n := range nutrients {
		q.Add("nutrients", fmt.Sprintf("%d", n))
	}

	var foods []Food
	if err := c.do(ctx, "GET", "v1/foods?"+q.Encode(), nil, &foods); err != nil {
		return nil, err
	}
	return foods, nil
}

// PostFoods retrieves multiple food items by FDC IDs (up to 20) using a JSON
// request body. This is the POST variant of GetFoods.
func (c *Client) PostFoods(ctx context.Context, criteria *FoodsCriteria) ([]Food, error) {
	if criteria == nil {
		return nil, fmt.Errorf("fdc: criteria is required")
	}

	var foods []Food
	if err := c.do(ctx, "POST", "v1/foods", criteria, &foods); err != nil {
		return nil, err
	}
	return foods, nil
}
