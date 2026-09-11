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

// MacroGateNumbers are the FDC nutrient numbers that must carry a non-zero
// per-100-g value for a candidate to pass the validity gate: energy,
// protein, total lipid, carbohydrate by difference.
var MacroGateNumbers = []uint{1008, 1003, 1004, 1005}

// macroGateKeys maps gate nutrient numbers to short per-100-g labels used in
// PrepareCandidate.Macros.
var macroGateKeys = map[uint]string{
	1008: "calories",
	1003: "protein",
	1004: "fat",
	1005: "carbohydrates",
}

// dataTypeRank implements the FDC selection priority: SR Legacy first, then
// Foundation, then Survey (FNDDS), Branded last. Unknown types rank last.
func dataTypeRank(dataType string) int {
	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "sr legacy":
		return 0
	case "foundation":
		return 1
	case "survey (fndds)", "survey":
		return 2
	case "branded":
		return 3
	default:
		return 4
	}
}

// PrepareCandidate is one USDA FDC search hit with the validity gate
// applied. Candidates are ranked by data type priority (SR Legacy,
// Foundation, Survey, Branded) then by search score.
type PrepareCandidate struct {
	FDCID       int                `json:"fdc_id"`
	Description string             `json:"description"`
	DataType    string             `json:"data_type"`
	Score       float64            `json:"score,omitempty"`
	MacroOK     bool               `json:"macro_ok"`
	Macros      map[string]float64 `json:"macros,omitempty"` // per-100-g values for the gate nutrients
}

// PrepareResult is the outcome of Prepare. Status is "candidates" (no
// write: the caller must pick an FDC ID and re-run) or "prepared" (the
// food exists with its FDC properties attached and conversions created).
type PrepareResult struct {
	Status       string                 `json:"status"`
	Name         string                 `json:"name"`
	Food         *food.Food             `json:"food,omitempty"`
	Created      bool                   `json:"created,omitempty"`
	FDCID        int                    `json:"fdc_id,omitempty"`
	Attached     []FdcAttachedProperty  `json:"attached,omitempty"`
	Conversions  *AutoConversionsResult `json:"conversions,omitempty"`
	Candidates   []PrepareCandidate     `json:"candidates,omitempty"`
	ActionsTaken []string               `json:"actions_taken"`
}

// PrepareOptions configures Prepare.
type PrepareOptions struct {
	// Name is the canonical food name (used for the exact-match lookup and
	// the FDC search query when the food is missing).
	Name string
	// FDCID, when set, is the chosen FDC ID. When nil, a missing food (or
	// an existing food without one) yields candidate mode: ranked FDC
	// candidates are returned without writing.
	FDCID *int
	// PluralName / Description are used only when the food is created.
	PluralName  string
	Description string
	// AutoConversions, when nil or true, runs AutoConversions after
	// attaching properties.
	AutoConversions *bool
	// Per100Amount / Per100UnitID are passed through to the FDC attach.
	Per100Amount float64
	Per100UnitID int
	// CandidateLimit caps candidate mode results (default 8, max 30).
	CandidateLimit int
}

// Prepare resolves a food by exact name, then brings it to a macro-complete
// state in one pass:
//
//   - food exists with an FDC ID (its own or the provided override) →
//     attach all FDC properties (idempotent) and create auto conversions.
//   - food exists with no FDC ID and none provided → candidate mode
//     (ranked FDC candidates, no write).
//   - food missing and fdc_id provided → create the food with the FDC ID,
//     then attach and convert.
//   - food missing and no fdc_id → candidate mode (ranked FDC candidates,
//     no write).
//
// Candidate mode never writes. The prepared path is idempotent: re-running
// updates existing properties and skips existing conversions.
func Prepare(ctx context.Context, c *tandoor.Client, fdcClient *fdc.Client, opts *PrepareOptions) (*PrepareResult, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if fdcClient == nil {
		return nil, fmt.Errorf("FDC API key not set")
	}
	res := &PrepareResult{Name: name, ActionsTaken: []string{}}

	existing, err := findExactByName(ctx, c, name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		res.Food = existing
		res.Created = false

		fdcID := 0
		if opts.FDCID != nil {
			fdcID = *opts.FDCID
		} else if existing.FDCID != nil {
			fdcID = *existing.FDCID
		}
		if fdcID == 0 {
			res.Status, res.Candidates, err = candidateMode(ctx, fdcClient, name, opts.CandidateLimit)
			if err != nil {
				return nil, err
			}
			res.ActionsTaken = append(res.ActionsTaken, "fdc_candidates_returned")
			return res, nil
		}

		// Apply the FDC ID override before attaching.
		if opts.FDCID != nil && (existing.FDCID == nil || *existing.FDCID != *opts.FDCID) {
			updated, err := c.Foods().Patch(ctx, &food.Food{ID: existing.ID, FDCID: opts.FDCID})
			if err != nil {
				return nil, fmt.Errorf("set fdc_id on food %d: %w", existing.ID, err)
			}
			res.Food = updated
			res.ActionsTaken = append(res.ActionsTaken, "fdc_id_set")
		}
		return finishPrepared(ctx, c, fdcClient, res, opts, fdcID)
	}

	// Food missing.
	if opts.FDCID == nil {
		res.Status, res.Candidates, err = candidateMode(ctx, fdcClient, name, opts.CandidateLimit)
		if err != nil {
			return nil, err
		}
		res.ActionsTaken = append(res.ActionsTaken, "fdc_candidates_returned")
		return res, nil
	}

	created, err := c.Foods().Create(ctx, &food.Food{
		Name:        name,
		PluralName:  opts.PluralName,
		Description: opts.Description,
		FDCID:       opts.FDCID,
	})
	if err != nil {
		return nil, fmt.Errorf("create food %q: %w", name, err)
	}
	res.Food = created
	res.Created = true
	res.ActionsTaken = append(res.ActionsTaken, "food_created")
	return finishPrepared(ctx, c, fdcClient, res, opts, *opts.FDCID)
}

// finishPrepared attaches the FDC properties and creates auto conversions
// for a food that has a resolved FDC ID.
func finishPrepared(ctx context.Context, c *tandoor.Client, fdcClient *fdc.Client,
	res *PrepareResult, opts *PrepareOptions, fdcID int,
) (*PrepareResult, error) {
	attachRes, err := AttachFDCProperties(ctx, c, fdcClient, &FdcAttachOptions{
		FoodID:       res.Food.ID,
		FDCID:        &fdcID,
		Per100Amount: opts.Per100Amount,
		Per100UnitID: opts.Per100UnitID,
	}, false)
	if err != nil {
		return nil, err
	}
	res.FDCID = fdcID
	res.Attached = attachRes.Attached
	res.ActionsTaken = append(res.ActionsTaken, attachRes.ActionsTaken...)

	if opts.AutoConversions == nil || *opts.AutoConversions {
		convRes, err := AutoConversions(ctx, c, res.Food.ID, false)
		if err != nil {
			return nil, err
		}
		res.Conversions = convRes
		res.ActionsTaken = append(res.ActionsTaken, convRes.ActionsTaken...)
	}

	res.Status = "prepared"
	return res, nil
}

// candidateMode searches FDC, ranks the hits (data type priority, then
// score), and applies the macro gate. It never writes.
func candidateMode(ctx context.Context, fdcClient *fdc.Client, name string, limit int) (string, []PrepareCandidate, error) {
	if limit <= 0 || limit > 30 {
		limit = 8
	}
	resp, err := fdcClient.SearchFoods(ctx, name, nil, &limit, nil, "", "", "")
	if err != nil {
		return "", nil, fmt.Errorf("FDC search for %q: %w", name, err)
	}

	candidates := make([]PrepareCandidate, 0, len(resp.Foods))
	for _, f := range resp.Foods {
		c := PrepareCandidate{
			FDCID:       f.FDCID,
			Description: f.Description,
			DataType:    f.DataType,
		}
		if f.Score != nil {
			c.Score = *f.Score
		}
		c.MacroOK, c.Macros = macroGate(f.FoodNutrients)
		candidates = append(candidates, c)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		ri, rj := dataTypeRank(candidates[i].DataType), dataTypeRank(candidates[j].DataType)
		if ri != rj {
			return ri < rj
		}
		return candidates[i].Score > candidates[j].Score
	})
	return "candidates", candidates, nil
}

// macroGate checks that every gate nutrient carries a non-zero amount and
// returns the per-100-g values found.
func macroGate(nutrients []fdc.AbridgedFoodNutrient) (bool, map[string]float64) {
	byNumber := make(map[uint]*float64, len(nutrients))
	for _, n := range nutrients {
		if n.Amount == nil {
			continue
		}
		v := *n.Amount
		byNumber[n.Number] = &v
	}
	ok := true
	macros := map[string]float64{}
	for _, num := range MacroGateNumbers {
		key, _ := macroGateKeys[num]
		if v, found := byNumber[num]; found && *v != 0 {
			macros[key] = *v
		} else {
			ok = false
		}
	}
	if !ok {
		macros = nil
	}
	return ok, macros
}

// findExactByName looks for a case-insensitive exact name match in a single
// search page (same lookup strategy as Ensure).
func findExactByName(ctx context.Context, c *tandoor.Client, name string) (*food.Food, error) {
	page, err := c.Foods().List(ctx, &food.ListOptions{
		ListOptions: pagination.ListOptions{Page: 1, PageSize: 100, Search: name},
	})
	if err != nil {
		return nil, fmt.Errorf("foods list for %q: %w", name, err)
	}
	for i := range page.Results {
		if strings.EqualFold(page.Results[i].Name, name) {
			return &page.Results[i], nil
		}
	}
	return nil, nil
}
