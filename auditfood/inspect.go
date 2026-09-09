package auditfood

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/detector"
	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/ingredient"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// FDCMatch is one USDA FDC search hit for a food without an FDC ID.
type FDCMatch struct {
	FDCID       int     `json:"fdc_id"`
	Description string  `json:"description"`
	DataType    string  `json:"data_type"`
	Score       float64 `json:"score,omitempty"`
}

// InspectResult is the deep-dive report for one food.
type InspectResult struct {
	Food          *food.Food              `json:"food"`
	Ingredients   []ingredient.Ingredient `json:"ingredients"`
	IssueCount    int                     `json:"issue_count"`
	Issues        []string                `json:"issues"`
	SuggestedName string                  `json:"suggested_name"`
	PrepNote      string                  `json:"prep_note,omitempty"`
	Alternatives  []string                `json:"alternatives,omitempty"`
	FDCMatches    []FDCMatch              `json:"fdc_matches"`
	Warnings      []string                `json:"warnings,omitempty"`
}

// Inspect reports a food's details, ingredient usage, detector issues, the
// suggested normalized name, and (when the food has no FDC ID and an FDC
// client is available) FDC candidates. The FDC client is optional.
func Inspect(ctx context.Context, c *tandoor.Client, fdcClient *fdc.Client, foodID, pageSize int) (*InspectResult, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	f, err := c.Foods().Get(ctx, foodID)
	if err != nil {
		return nil, fmt.Errorf("get food: %w", err)
	}

	// Ingredients using this food (server-side food filter).
	ingPage, err := c.Ingredients().List(ctx, &ingredient.ListOptions{
		ListOptions: pagination.ListOptions{
			PageSize: pageSize,
			Extra:    map[string]string{"food": fmt.Sprintf("%d", foodID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list ingredients: %w", err)
	}
	ingredients := make([]ingredient.Ingredient, len(ingPage.Results))
	copy(ingredients, ingPage.Results)

	// Detectors; warnings (e.g. property-type enumeration failure) are
	// collected rather than printed.
	var warnings []string
	warn := func(format string, args ...any) { warnings = append(warnings, fmt.Sprintf(format, args...)) }
	reg := NewRegistry(ctx, c, warn)
	detFood := detector.Food{
		Name:            f.Name,
		FDCID:           f.FDCID,
		PropertyTypeIDs: PropertyTypeIDs(f.Properties),
	}
	issues := reg.CheckAll(detFood)
	issueList := make([]string, len(issues))
	for i, iss := range issues {
		issueList[i] = fmt.Sprintf("[%s] %s (%s)", iss.Category, iss.Message, iss.Severity)
	}

	norm := normalize.Normalize(f.Name)

	var fdcMatches []FDCMatch
	if f.FDCID == nil && fdcClient != nil {
		limit := 5
		if resp, err := fdcClient.SearchFoods(ctx, norm.Canonical, nil, &limit, nil, "", "", ""); err == nil {
			for _, fd := range resp.Foods {
				m := FDCMatch{FDCID: fd.FDCID, Description: fd.Description, DataType: fd.DataType}
				if fd.Score != nil {
					m.Score = *fd.Score
				}
				fdcMatches = append(fdcMatches, m)
			}
		}
	}
	if fdcMatches == nil {
		fdcMatches = []FDCMatch{}
	}

	return &InspectResult{
		Food:          f,
		Ingredients:   ingredients,
		IssueCount:    len(issues),
		Issues:        issueList,
		SuggestedName: norm.Canonical,
		PrepNote:      norm.PrepNote,
		Alternatives:  norm.Alternatives,
		FDCMatches:    fdcMatches,
		Warnings:      warnings,
	}, nil
}
