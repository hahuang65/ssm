package errs_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hahuang65/ssm/errs"
)

var (
	errFoo = errors.New("foo sentinel")
	errBar = errors.New("bar sentinel")
)

func matchSubstring(needle string) func(error) bool {
	return func(err error) bool { return err != nil && strings.Contains(err.Error(), needle) }
}

func newMapper() *errs.Mapper {
	return errs.New(
		errs.Mapping{Sentinel: errFoo, Match: matchSubstring("foo"), Message: "foo hint"},
		errs.Mapping{Sentinel: errBar, Match: matchSubstring("bar"), Message: "bar hint"},
	)
}

func TestWrapNil(t *testing.T) {
	if got := newMapper().Wrap(nil); got != nil {
		t.Fatalf("Wrap(nil) = %v, want nil", got)
	}
}

func TestWrapMatched(t *testing.T) {
	raw := errors.New("the foo broke")
	got := newMapper().Wrap(raw)

	if !errors.Is(got, errFoo) {
		t.Fatalf("expected errors.Is(_, errFoo); got %v", got)
	}
	if !errors.Is(got, raw) {
		t.Fatalf("original error must remain reachable; got %v", got)
	}
}

func TestWrapUnmatchedPassesThrough(t *testing.T) {
	raw := errors.New("totally unrelated")
	got := newMapper().Wrap(raw)
	if got != raw {
		t.Fatalf("expected pass-through, got %v", got)
	}
}

func TestWrapOrderFirstWins(t *testing.T) {
	// "foobar" matches both mappings; foo is first, so foo wins.
	raw := errors.New("foobar exploded")
	got := newMapper().Wrap(raw)
	if !errors.Is(got, errFoo) {
		t.Fatalf("first-match should win; got %v", got)
	}
	if errors.Is(got, errBar) {
		t.Fatalf("second mapping should not fire; got %v", got)
	}
}

func TestMessageBySentinel(t *testing.T) {
	wrapped := fmt.Errorf("%w: underlying", errFoo)
	if got := newMapper().Message(wrapped); got != "foo hint" {
		t.Fatalf("Message = %q, want %q", got, "foo hint")
	}
}

func TestMessageByMatcherOnRawError(t *testing.T) {
	if got := newMapper().Message(errors.New("bar happened")); got != "bar hint" {
		t.Fatalf("Message = %q, want %q", got, "bar hint")
	}
}

func TestMessageFallsBackToError(t *testing.T) {
	if got := newMapper().Message(errors.New("nothing here")); got != "nothing here" {
		t.Fatalf("Message = %q, want fallback to err.Error()", got)
	}
}

func TestMessageNil(t *testing.T) {
	if got := newMapper().Message(nil); got != "" {
		t.Fatalf("Message(nil) = %q, want empty", got)
	}
}

func TestZeroMapperPassesThrough(t *testing.T) {
	var m *errs.Mapper
	raw := errors.New("boom")
	if got := m.Wrap(raw); got != raw {
		t.Fatalf("nil Mapper.Wrap should pass through, got %v", got)
	}
	if got := m.Message(raw); got != "boom" {
		t.Fatalf("nil Mapper.Message should fall back to err.Error(), got %q", got)
	}
}
