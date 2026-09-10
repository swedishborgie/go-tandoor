package pagination

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestNextPageNumber(t *testing.T) {
	p := &Paginated[int]{}
	assert.Equal(t, 0, p.NextPageNumber())

	p = &Paginated[int]{Next: strPtr("http://x/api/?page=3")}
	assert.Equal(t, 3, p.NextPageNumber())

	p = &Paginated[int]{Next: strPtr("http://x/api/?page=abc")}
	assert.Equal(t, 0, p.NextPageNumber())

	p = &Paginated[int]{Next: strPtr("http://x/api/")}
	assert.Equal(t, 0, p.NextPageNumber())
}

func TestPreviousPageNumber(t *testing.T) {
	p := &Paginated[int]{}
	assert.Equal(t, 0, p.PreviousPageNumber())

	p = &Paginated[int]{Previous: strPtr("http://x/api/?page=2")}
	assert.Equal(t, 2, p.PreviousPageNumber())
}

func TestListOptionsBuilder_OrderBy(t *testing.T) {
	assert.Equal(t, ListOptions{}, NewListOptions().OrderBy("", true).Build())
	assert.Equal(t, "name", NewListOptions().OrderBy("name", true).Build().OrderBy)
	assert.Equal(t, "-name", NewListOptions().OrderBy("name", false).Build().OrderBy)
}

func TestIterator_CurrentAndPage(t *testing.T) {
	it := &Iterator[int]{}
	var zero int
	assert.Equal(t, zero, it.Current())
	assert.Nil(t, it.Page())

	// Two pages of two items each.
	call := 0
	fetch := func() (*Paginated[int], error) {
		call++
		if call == 1 {
			return &Paginated[int]{Results: []int{1, 2}, Next: strPtr("http://x/?page=2")}, nil
		}
		return &Paginated[int]{Results: []int{3, 4}}, nil
	}
	it = NewIterator(fetch)
	var got []int
	for it.Next() {
		got = append(got, it.Current())
	}
	require.NoError(t, it.Err())
	assert.Equal(t, []int{1, 2, 3, 4}, got)
	assert.NotNil(t, it.Page())
	assert.Len(t, it.Page().Results, 2)
}

func TestIterator_FetchError(t *testing.T) {
	want := errors.New("boom")
	it := NewIterator(func() (*Paginated[int], error) { return nil, want })
	assert.False(t, it.Next())
	assert.Equal(t, want, it.Err())
}

func TestCollectAllExtra(t *testing.T) {
	ctx := context.Background()

	_, err := CollectAll(ctx, nil, func(int) (*Paginated[int], error) { return nil, nil })
	require.Error(t, err)

	// Multi-page: page 2's next URL has no page number, so CollectAll uses
	// the fallback (always requests page 2 again). The call counter makes
	// the second call return the final page so the loop terminates.
	first := &Paginated[int]{Results: []int{1}, Next: strPtr("http://x/?page=2")}
	page2 := &Paginated[int]{Results: []int{2}, Next: strPtr("http://x/no-page")}
	page3 := &Paginated[int]{Results: []int{3}}
	calls := 0
	got, err := CollectAll(ctx, first, func(page int) (*Paginated[int], error) {
		calls++
		if calls == 1 {
			assert.Equal(t, 2, page)
			return page2, nil
		}
		assert.Equal(t, 2, page) // fallback path: unparseable next URL -> 2
		return page3, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, got)
	assert.Equal(t, 2, calls)

	// Error on subsequent fetch.
	want := errors.New("boom")
	_, err = CollectAll(ctx, first, func(int) (*Paginated[int], error) {
		return nil, want
	})
	assert.Equal(t, want, err)

	// nil next response stops iteration.
	got, err = CollectAll(ctx, first, func(int) (*Paginated[int], error) { return nil, nil })
	require.NoError(t, err)
	assert.Equal(t, []int{1}, got)
}
