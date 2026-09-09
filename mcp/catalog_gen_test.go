package mcp

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	tandoor "github.com/swedishborgie/go-tandoor"
)

// This test keeps the MCP tool catalog in README.md in sync with the
// registered tools so the docs cannot drift:
//
//	go test ./mcp                              # verify the README catalog is current
//	go generate ./mcp                          # rewrite the README catalog
//	go test ./mcp -run TestCatalog -args -dump # print the catalog to stdout
//
// Custom flags must go after -args: the go command treats unknown flags
// before the package as its own, which silently changes the package list.
//
// The catalog is rendered in group order (the registration order of
// NewServer) with tools sorted alphabetically within each group.

var (
	genCatalog  = flag.Bool("gen", false, "rewrite the README tool catalog instead of verifying it")
	dumpCatalog = flag.Bool("dump", false, "print the tool catalog to stdout and skip verification")
)

const (
	catalogBegin = "<!-- BEGIN GENERATED: MCP tool catalog (go generate ./mcp) -->"
	catalogEnd   = "<!-- END GENERATED: MCP tool catalog -->"
)

// toolGroup is a catalog section: one human-readable title plus the tools
// registered by a single register* function.
type toolGroup struct {
	title string
	tools []toolDef
}

// allToolGroups reproduces the registration order of NewServer so the
// catalog groups match how the tools are organized in the source.
func allToolGroups() []toolGroup {
	d := &deps{Tandoor: mustTestClient()}
	return []toolGroup{
		{"Recipes", registerRecipeTools(d)},
		{"Ingredients", registerIngredientTools(d)},
		{"Steps", registerStepTools(d)},
		{"Foods", registerFoodTools(d)},
		{"Keywords", registerKeywordTools(d)},
		{"Units & Conversions", registerUnitTools(d)},
		{"Properties", registerPropertyTools(d)},
		{"Recipe Books", registerBookTools(d)},
		{"Shopping", registerShoppingTools(d)},
		{"Meal Plans", registerMealplanTools(d)},
		{"Cook Logs", registerCookLogTools(d)},
		{"Inventory", registerInventoryTools(d)},
		{"Storages", registerStorageTools(d)},
		{"Supermarkets", registerSupermarketTools(d)},
		{"Imports & Sharing", registerImportTools(d)},
		{"Syncs", registerSyncTools(d)},
		{"Spaces, Users & Groups", registerSpaceTools(d)},
		{"Auth", registerAuthTools(d)},
		{"Misc & Admin", registerMiscTools(d)},
		{"Audit Composites", registerAuditTools(d)},
		{"FoodData Central (FDC)", registerFdcTools(d)},
		{"Server", registerServerTools(d)},
	}
}

func mustTestClient() *tandoor.Client {
	c, err := tandoor.NewClient("http://127.0.0.1:1")
	if err != nil {
		// NewClient only parses the URL; this never happens in practice.
		panic(err)
	}
	return c
}

// renderCatalog renders the full generated block (markers included).
func renderCatalog() string {
	groups := allToolGroups()
	var b strings.Builder
	b.WriteString(catalogBegin + "\n\n")
	for i, g := range groups {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "#### %s\n\n", g.title)
		b.WriteString("| Tool | Description |\n| --- | --- |\n")
		sorted := append([]toolDef(nil), g.tools...)
		sort.Slice(sorted, func(j, k int) bool { return sorted[j].tool.Name < sorted[k].tool.Name })
		for _, td := range sorted {
			desc := strings.ReplaceAll(strings.TrimSpace(td.tool.Description), "|", `\|`)
			fmt.Fprintf(&b, "| `%s` | %s |\n", td.tool.Name, desc)
		}
	}
	b.WriteString("\n" + catalogEnd)
	return b.String()
}

func readmePath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "..", "README.md")
}

func extractCatalogBlock(readme string) (string, bool) {
	beginIdx := strings.Index(readme, catalogBegin)
	endIdx := strings.Index(readme, catalogEnd)
	if beginIdx < 0 || endIdx < 0 || endIdx < beginIdx {
		return "", false
	}
	return readme[beginIdx : endIdx+len(catalogEnd)], true
}

// TestCatalog verifies that the generated catalog in README.md matches the
// currently registered tools. Run `go generate ./mcp` after changing tool
// registrations or descriptions.
func TestCatalog(t *testing.T) {
	if *dumpCatalog {
		fmt.Print(renderCatalog())
		return
	}

	want := renderCatalog()
	readme := readmePath(t)
	data, err := os.ReadFile(readme)
	if err != nil {
		t.Skipf("cannot read %s: %v", readme, err)
	}
	got, ok := extractCatalogBlock(string(data))
	if !ok {
		t.Fatalf("README.md is missing the generated catalog block (%s … %s); run `go generate ./mcp`", catalogBegin, catalogEnd)
	}
	if *genCatalog {
		if got == want {
			t.Log("README catalog already up to date")
			return
		}
		replaced := string(data)
		bi := strings.Index(replaced, catalogBegin)
		ei := strings.Index(replaced, catalogEnd)
		replaced = replaced[:bi] + want + replaced[ei+len(catalogEnd):]
		if err := os.WriteFile(readme, []byte(replaced), 0o644); err != nil {
			t.Fatalf("writing %s: %v", readme, err)
		}
		t.Log("regenerated README catalog")
		return
	}
	if got != want {
		t.Fatalf("README.md tool catalog is out of date; run `go generate ./mcp`\n--- got (README) ---\n%s\n--- want (generated) ---\n%s", got, want)
	}
}
