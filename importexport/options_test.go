package importexport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestRecipeImportListOptions_Values(t *testing.T) {
	v := RecipeImportListOptions{ListOptions: pagination.ListOptions{Page: 2}}.Values()
	assert.Equal(t, "2", v.Get("page"))
}

func TestImportLogListOptions_Values(t *testing.T) {
	v := ImportLogListOptions{ListOptions: pagination.ListOptions{PageSize: 99}}.Values()
	assert.Equal(t, "99", v.Get("page_size"))
}

func TestExportLogListOptions_Values(t *testing.T) {
	v := ExportLogListOptions{ListOptions: pagination.ListOptions{OrderBy: "-date"}}.Values()
	assert.Equal(t, "-date", v.Get("ordering"))
}

func TestBookmarkletImportListOptions_Values(t *testing.T) {
	v := BookmarkletImportListOptions{ListOptions: pagination.ListOptions{Search: "x"}}.Values()
	assert.Equal(t, "x", v.Get("query"))
}

func TestSyncListOptions_Values(t *testing.T) {
	v := SyncListOptions{ListOptions: pagination.ListOptions{Page: 3}}.Values()
	assert.Equal(t, "3", v.Get("page"))
}

func TestSyncLogListOptions_Values(t *testing.T) {
	v := SyncLogListOptions{ListOptions: pagination.ListOptions{PageSize: 5}}.Values()
	assert.Equal(t, "5", v.Get("page_size"))
}
