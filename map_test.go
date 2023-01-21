package inithook_test

import (
	"testing"

	"github.com/ccmonky/inithook"
	"github.com/pkg/errors"
)

func TestError(t *testing.T) {
	err := errors.WithMessagef(inithook.ErrNotFound, "%d:%d", 1, 3)
	if !errors.Is(err, inithook.ErrNotFound) {
		t.Fatal("should is")
	}
}
