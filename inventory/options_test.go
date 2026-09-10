package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swedishborgie/go-tandoor/pagination"
)

func TestNewService(t *testing.T) {
	s := NewService(nil)
	assert.NotNil(t, s)
}

func TestLocationListOptions_Values(t *testing.T) {
	v := LocationListOptions{ListOptions: pagination.ListOptions{Page: 2}}.Values()
	assert.Equal(t, "2", v.Get("page"))
}
