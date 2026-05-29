// Package errs translates low-level errors into sentinel errors paired with
// user-facing messages.
//
// A Mapper holds an ordered list of Mappings. Each Mapping pairs a matcher
// predicate with a sentinel error and a human-readable message. Wrap turns a
// matching error into one that errors.Is-matches the sentinel while keeping
// the original error reachable via errors.As. Message returns the registered
// message for an error — found either by sentinel (if the error was wrapped)
// or by the matcher itself (so Message works on raw errors too).
//
// The package has no domain knowledge; callers compose Mappings for whatever
// error space they care about.
package errs

import (
	"errors"
	"fmt"
)

// Mapping pairs a matcher with a sentinel and a user-facing message.
type Mapping struct {
	// Sentinel is wrapped into matching errors so callers can detect them
	// with errors.Is.
	Sentinel error
	// Match reports whether an error chain belongs to this mapping. It is
	// called on every Wrap/Message call, so it should be cheap.
	Match func(error) bool
	// Message is the user-facing text returned by Mapper.Message when the
	// mapping fires.
	Message string
}

// Mapper resolves errors against an ordered list of mappings. The zero value
// is a valid empty Mapper that passes errors through unchanged.
type Mapper struct {
	mappings []Mapping
}

// New returns a Mapper populated with the given mappings, evaluated in order;
// the first match wins.
func New(mappings ...Mapping) *Mapper {
	return &Mapper{mappings: mappings}
}

// Wrap returns err wrapped with the matching sentinel and the original error
// when a mapping fires, or err unchanged otherwise. Wrap(nil) returns nil.
func (m *Mapper) Wrap(err error) error {
	if err == nil || m == nil {
		return err
	}
	for _, mp := range m.mappings {
		if mp.Match != nil && mp.Match(err) {
			return fmt.Errorf("%w: %w", mp.Sentinel, err)
		}
	}
	return err
}

// Message returns the user-facing text registered for err, looked up by
// sentinel first (so wrapped errors short-circuit) and by matcher second (so
// raw errors still resolve). When nothing matches, Message returns err.Error().
// Message(nil) returns the empty string.
func (m *Mapper) Message(err error) string {
	if err == nil {
		return ""
	}
	if m != nil {
		for _, mp := range m.mappings {
			if mp.Sentinel != nil && errors.Is(err, mp.Sentinel) {
				return mp.Message
			}
			if mp.Match != nil && mp.Match(err) {
				return mp.Message
			}
		}
	}
	return err.Error()
}
