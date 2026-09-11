package auditfood

import (
	"context"
	"fmt"

	"github.com/swedishborgie/go-tandoor"
	"github.com/swedishborgie/go-tandoor/detector"
	"github.com/swedishborgie/go-tandoor/pagination"
	"github.com/swedishborgie/go-tandoor/property"
	"github.com/swedishborgie/go-tandoor/unit"
)

// FetchAllUnits enumerates every unit on the server (auto-paginated).
func FetchAllUnits(ctx context.Context, c *tandoor.Client) ([]unit.Unit, error) {
	svc := c.Units()
	page, err := svc.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list units: %w", err)
	}
	return pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[unit.Unit], error) {
		opts := &unit.ListOptions{ListOptions: pagination.ListOptions{Page: pageNum, PageSize: 100}}
		return svc.List(ctx, opts)
	})
}

// FindGramUnitID resolves the instance's gram unit (see findGramUnitID).
func FindGramUnitID(ctx context.Context, c *tandoor.Client) (int, error) {
	return findGramUnitID(ctx, c)
}

// FetchAllPropertyTypes enumerates every property type on the server
// (auto-paginated).
func FetchAllPropertyTypes(ctx context.Context, c *tandoor.Client) ([]property.Type, error) {
	svc := c.PropertyTypes()
	page, err := svc.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list property types: %w", err)
	}
	return pagination.CollectAll(ctx, page, func(pageNum int) (*pagination.Paginated[property.Type], error) {
		return svc.List(ctx, &property.TypeListOptions{ListOptions: pagination.ListOptions{Page: pageNum}})
	})
}

// NewRegistry returns a detector registry with the built-in detectors plus
// missing_properties, whose expected type set is enumerated from the server.
// If the enumeration fails, warn is called (when non-nil) and the registry
// is returned without that detector — the audit's core job (names/FDC)
// still runs. Callers should build one registry per operation, not per food.
func NewRegistry(ctx context.Context, c *tandoor.Client, warn func(format string, args ...any)) *detector.Registry {
	reg := detector.NewRegistry()
	types, err := FetchAllPropertyTypes(ctx, c)
	if err != nil {
		if warn != nil {
			warn("audit: could not enumerate property types from server, skipping missing_properties check: %v", err)
		}
		return reg
	}
	expected := make([]detector.PropertyTypeInfo, 0, len(types))
	for _, t := range types {
		expected = append(expected, detector.PropertyTypeInfo{ID: t.ID, Name: t.Name})
	}
	reg.Add(detector.NewMissingPropertiesDetector(expected))
	return reg
}

// PropertyTypeIDs returns the ids of property types attached to the food.
func PropertyTypeIDs(props []property.Property) []int {
	ids := make([]int, 0, len(props))
	for _, p := range props {
		ids = append(ids, p.Type.ID)
	}
	return ids
}
