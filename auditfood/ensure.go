package auditfood

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/fdc"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// FdcCandidate is one USDA FDC search hit, for agent selection.
type FdcCandidate struct {
	FDCID       int     `json:"fdc_id"`
	Description string  `json:"description"`
	DataType    string  `json:"data_type"`
	Score       float64 `json:"score,omitempty"`
}

// NearMatch is an existing food close to a requested name.
type NearMatch struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Similarity float64 `json:"similarity"`
}

// EnsureResult is the outcome for one requested name. Status is one of
// "found", "created", "would_create", "ambiguous", "not_found".
type EnsureResult struct {
	Status        string         `json:"status"`
	Food          *food.Food     `json:"food,omitempty"`
	NearMatches   []NearMatch    `json:"near_matches,omitempty"`
	FdcCandidates []FdcCandidate `json:"fdc_candidates,omitempty"`
	ActionsTaken  []string       `json:"actions_taken"`
}

// EnsureSummary counts results by status.
type EnsureSummary struct {
	Total       int `json:"total"`
	Found       int `json:"found"`
	Created     int `json:"created"`
	WouldCreate int `json:"would_create"`
	Ambiguous   int `json:"ambiguous"`
	NotFound    int `json:"not_found"`
}

// EnsureReport is the full result of Ensure.
type EnsureReport struct {
	Summary EnsureSummary            `json:"summary"`
	Results map[string]*EnsureResult `json:"results"`
}

// EnsureOptions configures Ensure.
type EnsureOptions struct {
	Names []string
	// FDC is optional; when nil, FDC candidates are skipped (no API key
	// required to run ensure).
	FDC *fdc.Client
	// FDCLimit caps FDC candidates per name (default 5, max 200).
	FDCLimit int
	// Threshold is the Jaccard similarity floor for near matches (default 0.6).
	Threshold float64
	// ForceCreate creates the food when no exact match exists instead of
	// reporting it.
	ForceCreate bool
	// PageSize is the food search page size (default 100).
	PageSize int
}

func (o *EnsureOptions) defaults() {
	if o.FDCLimit <= 0 || o.FDCLimit > 200 {
		o.FDCLimit = 5
	}
	if o.Threshold <= 0 {
		o.Threshold = 0.6
	}
	if o.PageSize <= 0 {
		o.PageSize = 100
	}
}

// Ensure resolves each requested name against the instance: exact match
// (case-insensitive) wins; otherwise near matches (Jaccard >= Threshold) and
// FDC candidates are reported. With ForceCreate and no exact match, the food
// is created (or reported as would_create when dryRun).
func Ensure(ctx context.Context, c *tandoor.Client, opts *EnsureOptions, dryRun bool) (*EnsureReport, error) {
	opts.defaults()
	results := make(map[string]*EnsureResult, len(opts.Names))

	for _, name := range opts.Names {
		res := &EnsureResult{Status: "not_found", ActionsTaken: []string{}}

		page, err := c.Foods().List(ctx, &food.ListOptions{
			ListOptions: pagination.ListOptions{
				PageSize: opts.PageSize,
				Search:   name,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("foods list for %q: %w", name, err)
		}

		var exact *food.Food
		var candidates []food.Food
		for i := range page.Results {
			f := &page.Results[i]
			if strings.EqualFold(f.Name, name) {
				exact = f
				break
			}
			candidates = append(candidates, *f)
		}

		if exact != nil {
			res.Status = "found"
			res.Food = exact
			results[name] = res
			continue
		}

		near := []NearMatch{}
		for _, f := range candidates {
			sim := JaccardSimilarity(name, f.Name)
			if sim >= opts.Threshold {
				near = append(near, NearMatch{ID: f.ID, Name: f.Name, Similarity: sim})
			}
		}
		sort.Slice(near, func(i, j int) bool { return near[j].Similarity > near[i].Similarity })

		if len(near) > 0 {
			res.Status = "ambiguous"
		}
		res.NearMatches = near

		if opts.FDC != nil {
			res.FdcCandidates, err = fdcCandidates(ctx, opts.FDC, name, opts.FDCLimit)
			if err != nil {
				// FDC candidates are advisory — a lookup failure must not
				// fail the ensure.
				res.FdcCandidates = nil
			}
		}
		if res.FdcCandidates == nil {
			res.FdcCandidates = []FdcCandidate{}
		}

		if opts.ForceCreate {
			if dryRun {
				res.Status = "would_create"
				res.ActionsTaken = append(res.ActionsTaken, "food_create_dry_run")
			} else {
				created, err := c.Foods().Create(ctx, &food.Food{Name: name})
				if err != nil {
					return nil, fmt.Errorf("create food %q: %w", name, err)
				}
				res.Status = "created"
				res.Food = created
				res.ActionsTaken = append(res.ActionsTaken, "food_created")
			}
		}
		results[name] = res
	}

	summary := EnsureSummary{Total: len(opts.Names)}
	for _, r := range results {
		switch r.Status {
		case "found":
			summary.Found++
		case "created":
			summary.Created++
		case "would_create":
			summary.WouldCreate++
		case "ambiguous":
			summary.Ambiguous++
		default:
			summary.NotFound++
		}
	}
	return &EnsureReport{Summary: summary, Results: results}, nil
}

func fdcCandidates(ctx context.Context, client *fdc.Client, name string, limit int) ([]FdcCandidate, error) {
	pageSize := limit
	resp, err := client.SearchFoods(ctx, name, nil, &pageSize, nil, "", "", "")
	if err != nil {
		return nil, err
	}
	out := make([]FdcCandidate, 0, limit)
	for _, f := range resp.Foods {
		if len(out) >= limit {
			break
		}
		c := FdcCandidate{FDCID: f.FDCID, Description: f.Description, DataType: f.DataType}
		if f.Score != nil {
			c.Score = *f.Score
		}
		out = append(out, c)
	}
	return out, nil
}

// FindByName returns foods whose name equals name (case-insensitive),
// paginating the query search to completion.
func FindByName(ctx context.Context, c *tandoor.Client, name string, pageSize int) ([]food.Food, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	matches := make([]food.Food, 0)
	pageNum := 1
	for {
		page, err := c.Foods().List(ctx, &food.ListOptions{
			ListOptions: pagination.ListOptions{
				Page:     pageNum,
				PageSize: pageSize,
				Extra:    map[string]string{"query": name},
			},
		})
		if err != nil {
			return nil, err
		}
		for _, f := range page.Results {
			if strings.EqualFold(f.Name, name) {
				matches = append(matches, f)
			}
		}
		if !page.HasNext() {
			break
		}
		pageNum++
	}
	return matches, nil
}
