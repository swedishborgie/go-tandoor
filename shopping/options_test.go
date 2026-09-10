package shopping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestListListOptions_Values(t *testing.T) {
	v := ListListOptions{ListOptions: pagination.ListOptions{Search: "x", Page: 2}}.Values()
	assert.Equal(t, "x", v.Get("query"))
	assert.Equal(t, "2", v.Get("page"))
}

func TestListEntryListOptions_Values(t *testing.T) {
	v := ListEntryListOptions{ListOptions: pagination.ListOptions{PageSize: 50}}.Values()
	assert.Equal(t, "50", v.Get("page_size"))
}

func TestListRecipeListOptions_Values(t *testing.T) {
	v := ListRecipeListOptions{ListOptions: pagination.ListOptions{OrderBy: "-date"}}.Values()
	assert.Equal(t, "-date", v.Get("ordering"))
}
