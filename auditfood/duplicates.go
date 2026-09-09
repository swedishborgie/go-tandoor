package auditfood

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/swedishborgie/go-tandoor/pagination"
)

// DefaultDuplicateThreshold is the Jaccard similarity floor for duplicate
// grouping.
const DefaultDuplicateThreshold = 0.7

// DuplicateGroup is a set of foods similar enough to be the same food.
type DuplicateGroup struct {
	Foods []DuplicateFood `json:"foods"`
}

// DuplicateFood is one member of a duplicate group.
type DuplicateFood struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DuplicatesResult reports the duplicate scan.
type DuplicatesResult struct {
	Scanned   int              `json:"scanned"`
	Threshold float64          `json:"threshold"`
	Groups    []DuplicateGroup `json:"groups"`
}

// FindDuplicates scans all foods, normalizes each name, and groups foods
// whose normalized names are above the Jaccard threshold (transitive
// closure via union-find).
func FindDuplicates(ctx context.Context, c *tandoor.Client, threshold float64) (*DuplicatesResult, error) {
	if threshold <= 0 {
		threshold = DefaultDuplicateThreshold
	}
	page, err := c.Foods().List(ctx, &food.ListOptions{
		ListOptions: pagination.ListOptions{PageSize: 50},
	})
	if err != nil {
		return nil, fmt.Errorf("list foods: %w", err)
	}
	foods, err := pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[food.Food], error) {
		return c.Foods().List(ctx, &food.ListOptions{ListOptions: pagination.ListOptions{Page: pageNum, PageSize: 50}})
	})
	if err != nil {
		return nil, err
	}

	// Normalize and word-set each name.
	type entry struct {
		food  food.Food
		set   map[string]struct{}
		words []string
	}
	entries := make([]entry, 0, len(foods))
	for _, f := range foods {
		norm := normalize.Normalize(f.Name)
		entries = append(entries, entry{
			food:  f,
			set:   WordSet(strings.ToLower(norm.Canonical)),
			words: strings.Fields(strings.ToLower(norm.Canonical)),
		})
	}

	// Union-find over similar pairs.
	parent := make([]int, len(entries))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}
	union := func(i, j int) {
		ri, rj := find(i), find(j)
		if ri != rj {
			parent[rj] = ri
		}
	}
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			sim := Jaccard(entries[i].set, entries[j].set)
			if sim >= threshold {
				union(i, j)
			}
		}
	}

	// Collect groups (only groups with 2+ members).
	groupsByRoot := make(map[int][]entry)
	for i, e := range entries {
		groupsByRoot[find(i)] = append(groupsByRoot[find(i)], e)
	}
	groups := make([]DuplicateGroup, 0, len(groupsByRoot))
	for _, members := range groupsByRoot {
		if len(members) < 2 {
			continue
		}
		sort.Slice(members, func(i, j int) bool { return members[i].food.ID < members[j].food.ID })
		g := DuplicateGroup{}
		for _, m := range members {
			g.Foods = append(g.Foods, DuplicateFood{ID: m.food.ID, Name: m.food.Name})
		}
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Foods[0].ID < groups[j].Foods[0].ID
	})

	return &DuplicatesResult{Scanned: len(foods), Threshold: threshold, Groups: groups}, nil
}
