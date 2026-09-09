package auditfood

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/food"
	"github.com/swedishborgie/go-tandoor/normalize"
)

// FixOptions configures Fix.
type FixOptions struct {
	FoodID int
	// MergeOnCollision merges the food into an existing food with the same
	// normalized name instead of failing (default true). When false, a
	// collision is an error.
	MergeOnCollision bool
}

// FixResult reports the applied (or planned) name fix.
type FixResult struct {
	FoodID       int        `json:"food_id"`
	OriginalName string     `json:"original_name"`
	NewName      string     `json:"new_name"`
	PrepNote     string     `json:"prep_note,omitempty"`
	Alternatives []string   `json:"alternatives,omitempty"`
	Collision    *food.Food `json:"collision,omitempty"`
	Merged       bool       `json:"merged,omitempty"`
	Food         *food.Food `json:"food,omitempty"` // updated food, or the merge target
	ActionsTaken []string   `json:"actions_taken"`
}

// Fix renames a food to its normalized canonical name. Before writing, it
// checks for a name collision: if another food already has the canonical
// name, the fix merges into it (MergeOnCollision, default) or fails.
// dryRun returns the plan without writing.
func Fix(ctx context.Context, c *tandoor.Client, opts *FixOptions, dryRun bool) (*FixResult, error) {
	f, err := c.Foods().Get(ctx, opts.FoodID)
	if err != nil {
		return nil, fmt.Errorf("get food: %w", err)
	}
	norm := normalize.Normalize(f.Name)

	res := &FixResult{
		FoodID:       opts.FoodID,
		OriginalName: f.Name,
		NewName:      norm.Canonical,
		PrepNote:     norm.PrepNote,
		Alternatives: norm.Alternatives,
		ActionsTaken: []string{},
	}
	if norm.Canonical == f.Name {
		res.ActionsTaken = append(res.ActionsTaken, "name_already_canonical")
		return res, nil
	}

	// Collision check: another food with the exact canonical name.
	collisions, err := FindByName(ctx, c, norm.Canonical, 100)
	if err != nil {
		return nil, fmt.Errorf("collision check: %w", err)
	}
	var collision *food.Food
	for i := range collisions {
		if collisions[i].ID != f.ID {
			collision = &collisions[i]
			break
		}
	}
	if collision != nil {
		res.Collision = collision
		if dryRun {
			if opts.MergeOnCollision {
				res.ActionsTaken = append(res.ActionsTaken, "would_merge_into_"+fmt.Sprint(collision.ID))
			} else {
				res.ActionsTaken = append(res.ActionsTaken, "would_fail_collision")
			}
			return res, nil
		}
		if !opts.MergeOnCollision {
			return res, fmt.Errorf("name collision: food %d is already named %q (set merge=true to merge into it)", collision.ID, collision.Name)
		}
		merged, err := c.Foods().Merge(ctx, f.ID, collision.ID)
		if err != nil {
			return res, fmt.Errorf("merge food %d into %d: %w", f.ID, collision.ID, err)
		}
		res.Merged = true
		res.Food = merged
		res.ActionsTaken = append(res.ActionsTaken, fmt.Sprintf("merged_into_%d", collision.ID))
		return res, nil
	}

	if dryRun {
		res.ActionsTaken = append(res.ActionsTaken, "would_rename")
		return res, nil
	}

	updated, err := c.Foods().Patch(ctx, &food.Food{ID: f.ID, Name: norm.Canonical})
	if err != nil {
		return nil, fmt.Errorf("patch food %d: %w", opts.FoodID, err)
	}
	res.Food = updated
	res.ActionsTaken = append(res.ActionsTaken, "renamed")
	return res, nil
}
