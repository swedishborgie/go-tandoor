package pagination

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginated_HasNext(t *testing.T) {
	empty := ""
	p := Paginated[int]{
		Count:   100,
		Next:    &empty,
		Results: []int{},
	}
	assert.False(t, p.HasNext())

	url := "https://example.com/api/?page=2"
	p.Next = &url
	assert.True(t, p.HasNext())
}

func TestPaginated_HasPrevious(t *testing.T) {
	p := Paginated[int]{}
	assert.False(t, p.HasPrevious())

	url := "https://example.com/api/?page=1"
	p.Previous = &url
	assert.True(t, p.HasPrevious())
}

func TestPaginated_PageNumbers(t *testing.T) {
	nextURL := "https://example.com/api/recipe/?page=3"
	prevURL := "https://example.com/api/recipe/?page=1"

	p := Paginated[int]{
		Next:     &nextURL,
		Previous: &prevURL,
	}

	assert.Equal(t, 3, p.NextPageNumber())
	assert.Equal(t, 1, p.PreviousPageNumber())
}

func TestPaginated_PageNumbers_MissingPage(t *testing.T) {
	noParam := "https://example.com/api/recipe/?q=foo"
	invalid := "https://example.com/api/recipe/?page=abc"
	badURL := "not a url"

	nextURL := noParam
	p := Paginated[int]{Next: &nextURL}
	assert.Equal(t, 0, p.NextPageNumber())

	nextURL = invalid
	p.Next = &nextURL
	assert.Equal(t, 0, p.NextPageNumber())

	nextURL = badURL
	p.Next = &nextURL
	assert.Equal(t, 0, p.NextPageNumber())
}

func TestListOptions_Values(t *testing.T) {
	t.Run("empty options returns empty values", func(t *testing.T) {
		o := ListOptions{}
		assert.Empty(t, o.Values())
	})

	t.Run("page and page_size", func(t *testing.T) {
		o := ListOptions{Page: 2, PageSize: 25}
		v := o.Values()
		assert.Equal(t, "2", v.Get("page"))
		assert.Equal(t, "25", v.Get("page_size"))
	})

	t.Run("search", func(t *testing.T) {
		o := ListOptions{Search: "chicken"}
		v := o.Values()
		assert.Equal(t, "chicken", v.Get("query"))
	})

	t.Run("extra params", func(t *testing.T) {
		o := ListOptions{
			Extra: map[string]string{"keywords__name": "dinner"},
		}
		v := o.Values()
		assert.Equal(t, "dinner", v.Get("keywords__name"))
	})

	t.Run("combined", func(t *testing.T) {
		o := ListOptions{
			Page:     3,
			PageSize: 10,
			Search:   "pasta",
			Extra:    map[string]string{"ordering": "name"},
		}
		v := o.Values()
		assert.Equal(t, "3", v.Get("page"))
		assert.Equal(t, "10", v.Get("page_size"))
		assert.Equal(t, "pasta", v.Get("query"))
		assert.Equal(t, "name", v.Get("ordering"))
	})

	t.Run("empty extra values are skipped", func(t *testing.T) {
		o := ListOptions{
			Extra: map[string]string{"ignored": ""},
		}
		assert.Empty(t, o.Values())
	})
}

func TestListOptions_QueryString(t *testing.T) {
	o := ListOptions{Page: 2, Search: "test"}
	assert.Equal(t, "page=2&query=test", o.QueryString())

	oEmpty := ListOptions{}
	assert.Empty(t, oEmpty.QueryString())
}

func TestListOptionsBuilder(t *testing.T) {
	opts := NewListOptions().
		Page(3).
		PageSize(20).
		Query("pasta").
		OrderBy("name", false).
		WithExtra("ordering", "should be overridden")
	built := opts.Build()
	assert.Equal(t, 3, built.Page)
	assert.Equal(t, 20, built.PageSize)
	assert.Equal(t, "pasta", built.Search)
	assert.Equal(t, "-name", built.OrderBy)
	// WithExtra should not override OrderBy set by OrderBy method
	assert.Equal(t, "should be overridden", built.Extra["ordering"])
}

func TestIterator(t *testing.T) {
	type item struct {
		ID   int
		Name string
	}

	page := 1
	fetch := func() (*Paginated[item], error) {
		page++
		switch page {
		case 2:
			next := "https://example.com/api/?page=3"
			return &Paginated[item]{
				Count: 4,
				Next:  &next,
				Results: []item{
					{ID: 1, Name: "a"},
					{ID: 2, Name: "b"},
				},
			}, nil
		case 3:
			return &Paginated[item]{
				Count:    4,
				Results:  []item{{ID: 3, Name: "c"}, {ID: 4, Name: "d"}},
				Previous: stringPtr("https://example.com/api/?page=1"),
			}, nil
		default:
			return nil, assert.AnError
		}
	}

	iter := NewIterator(fetch)

	var ids []int
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []int{1, 2, 3, 4}, ids)
}

func TestIterator_Error(t *testing.T) {
	fetch := func() (*Paginated[int], error) {
		return nil, assert.AnError
	}

	iter := NewIterator(fetch)
	assert.False(t, iter.Next())
	assert.Equal(t, assert.AnError, iter.Err())
}

func TestCollectAll(t *testing.T) {
	type item struct {
		ID int
	}

	nextURL := "https://example.com/api/?page=2"
	first := &Paginated[item]{
		Count:   3,
		Next:    &nextURL,
		Results: []item{{ID: 1}, {ID: 2}},
	}

	callCount := 0
	nextFn := func(page int) (*Paginated[item], error) {
		callCount++
		if page == 2 {
			return &Paginated[item]{
				Count:   3,
				Results: []item{{ID: 3}},
			}, nil
		}
		return nil, nil
	}

	all, err := CollectAll(context.Background(), first, nextFn)
	require.NoError(t, err)
	assert.Equal(t, []item{{ID: 1}, {ID: 2}, {ID: 3}}, all)
	assert.Equal(t, 1, callCount)
}

func TestCollectAll_SinglePage(t *testing.T) {
	type item struct{ ID int }
	first := &Paginated[item]{
		Count:   2,
		Results: []item{{ID: 1}, {ID: 2}},
	}

	all, err := CollectAll(context.Background(), first, func(page int) (*Paginated[item], error) {
		return nil, assert.AnError
	})
	require.NoError(t, err)
	assert.Equal(t, []item{{ID: 1}, {ID: 2}}, all)
}

func stringPtr(s string) *string {
	return &s
}
