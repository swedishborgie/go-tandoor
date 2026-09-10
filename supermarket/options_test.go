package supermarket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestListOptions_Values(t *testing.T) {
	v := ListOptions{ListOptions: pagination.ListOptions{Page: 2}}.Values()
	assert.Equal(t, "2", v.Get("page"))
}

func TestListOptions_ToPaginationOptions(t *testing.T) {
	p := ListOptions{ListOptions: pagination.ListOptions{PageSize: 60}}.ToPaginationOptions()
	assert.Equal(t, 60, p.PageSize)
}
