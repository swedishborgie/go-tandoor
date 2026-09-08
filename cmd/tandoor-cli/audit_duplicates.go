// cmd/tandoor-cli/cmd_audit_duplicates.go

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/normalize"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/urfave/cli/v3"
)

// auditDuplicatesCommand returns the `audit duplicates` subcommand.
func auditDuplicatesCommand() *cli.Command {
	return &cli.Command{
		Name:  "duplicates",
		Usage: "Find near-duplicate foods by name similarity",
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:  "threshold",
				Usage: "Similarity threshold (0.0-1.0, lower = more matches)",
				Value: 0.7,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c := ctx.Value(ctxKeyClient).(*tandoor.Client)
			threshold := cmd.Float64("threshold")
			pageSize := cmd.Int("page-size")

			type FoodEntry struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}

			// 1. Fetch all foods and normalize names.
			var foods []FoodEntry
			opts := &food.ListOptions{ListOptions: pagination.ListOptions{Page: 1, PageSize: pageSize}}
			firstPage, err := c.Foods().List(ctx, opts)
			if err != nil {
				return fmt.Errorf("list foods page 1: %w", err)
			}
			allFoods, err := pagination.CollectAll(ctx, firstPage, func(pageNum int) (*pagination.Paginated[food.Food], error) {
				opts.Page = pageNum
				return c.Foods().List(ctx, opts)
			})
			if err != nil {
				return err
			}
			for _, f := range allFoods {
				foods = append(foods, FoodEntry{ID: f.ID, Name: f.Name})
			}

			// 2. Normalize each food name to a word set.
			type normalizedFood struct {
				id        int
				name      string
				canonical string
				words     map[string]struct{}
			}
			normalized := make([]normalizedFood, 0, len(foods))
			for _, f := range foods {
				norm := normalize.Normalize(f.Name)
				words := wordSet(norm.Canonical)
				if len(words) == 0 {
					continue // skip empty names
				}
				normalized = append(normalized, normalizedFood{
					id:        f.ID,
					name:      f.Name,
					canonical: norm.Canonical,
					words:     words,
				})
			}

			// 3. Compare all pairs with Jaccard similarity and build groups.
			parent := make(map[int]int) // union-find parent
			for i := 0; i < len(normalized); i++ {
				parent[i] = i
			}
			var findFn func(int) int
			findFn = func(x int) int {
				if parent[x] != x {
					parent[x] = findFn(parent[x])
				}
				return parent[x]
			}
			union := func(a, b int) {
				ra, rb := findFn(a), findFn(b)
				if ra != rb {
					parent[ra] = rb
				}
			}

			for i := 0; i < len(normalized); i++ {
				for j := i + 1; j < len(normalized); j++ {
					if jaccard(normalized[i].words, normalized[j].words) >= threshold {
						union(i, j)
					}
				}
			}

			// 4. Collect groups (only groups with 2+ members).
			groups := make(map[int][]normalizedFood)
			for i, nf := range normalized {
				root := findFn(i)
				groups[root] = append(groups[root], nf)
			}

			type DuplicateGroup struct {
				Group               []FoodEntry `json:"group"`
				SuggestedCanonical  string      `json:"suggested_canonical"`
				Action              string      `json:"action"`
				SimilarityThreshold float64     `json:"similarity_threshold"`
			}

			var results []DuplicateGroup
			for _, members := range groups {
				if len(members) < 2 {
					continue
				}
				// Sort by ID for deterministic output.
				for i := 0; i < len(members); i++ {
					for j := i + 1; j < len(members); j++ {
						if members[j].id < members[i].id {
							members[i], members[j] = members[j], members[i]
						}
					}
				}

				entries := make([]FoodEntry, len(members))
				for i, m := range members {
					entries[i] = FoodEntry{ID: m.id, Name: m.name}
				}

				// Suggested canonical = the shortest normalized name (most stripped).
				best := members[0].canonical
				for _, m := range members[1:] {
					if len(m.canonical) < len(best) {
						best = m.canonical
					}
				}

				results = append(results, DuplicateGroup{
					Group:               entries,
					SuggestedCanonical:  best,
					Action:              "merge",
					SimilarityThreshold: threshold,
				})
			}

			if results == nil {
				results = []DuplicateGroup{}
			}

			fmt.Fprintf(cmd.ErrWriter, "scanned %d foods, found %d duplicate groups (threshold %.2f)\n",
				len(normalized), len(results), threshold)

			return printJSON(results)
		},
	}
}

// wordSet splits a canonical name into a set of lowercased words.
func wordSet(name string) map[string]struct{} {
	words := strings.Fields(strings.ToLower(name))
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		set[w] = struct{}{}
	}
	return set
}

// jaccard computes Jaccard similarity between two word sets.
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	intersection := 0
	for w := range a {
		if _, ok := b[w]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
