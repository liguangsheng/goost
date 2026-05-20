package taskgroup

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResultsCollectAll(t *testing.T) {
	g := NewResults[int](context.Background())
	for i := range 5 {
		g.Run(func(_ context.Context) (int, error) {
			return i * i, nil
		})
	}
	values, err := g.Wait()
	assert.NoError(t, err)
	sort.Ints(values)
	assert.Equal(t, []int{0, 1, 4, 9, 16}, values)
}

func Test_ResultsFirstError(t *testing.T) {
	g := NewResults[int](context.Background())
	fail := errors.New("fail")
	g.Run(func(_ context.Context) (int, error) { return 0, fail })
	g.Run(func(_ context.Context) (int, error) { return 1, nil })
	_, err := g.Wait()
	assert.ErrorIs(t, err, fail)
}

func Test_ResultsLimit(t *testing.T) {
	g := NewResults[int](context.Background()).WithLimit(2)
	for i := range 10 {
		g.Run(func(_ context.Context) (int, error) { return i, nil })
	}
	values, err := g.Wait()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(values))
}

func Test_ResultsPanic(t *testing.T) {
	g := NewResults[int](context.Background())
	g.Run(func(_ context.Context) (int, error) { panic("oops") })
	_, err := g.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic")
}
