package inithook_test

import (
	"context"
	"testing"

	"github.com/ccmonky/inithook"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := errors.WithMessagef(inithook.ErrNotFound, "%d:%d", 1, 3)
	if !errors.Is(err, inithook.ErrNotFound) {
		t.Fatal("should is")
	}
}

func TestMap(t *testing.T) {
	m := inithook.NewMap[int, string]()
	ctx := context.Background()
	err := m.Set(ctx, 1, "one")
	assert.Nilf(t, err, "set one")
	err = m.Set(ctx, 2, "two")
	assert.Nilf(t, err, "set one")
	assert.ElementsMatchf(t, []int{1, 2}, m.Keys(ctx), "keys")
	assert.ElementsMatchf(t, []string{"one", "two"}, m.Values(ctx), "value")
	assert.Equalf(t, map[int]string{
		1: "one",
		2: "two",
	}, m.Map(ctx), "map")
}
