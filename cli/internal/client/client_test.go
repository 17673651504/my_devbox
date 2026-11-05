package client

import (
	"errors"
	"net/http"
	"testing"
)

// TestAPIErrorIsUnauthorized verifies that a 401 APIError matches the
// ErrUnauthorized sentinel via errors.Is, so all unauthorized failures are
// recognized uniformly no matter which code path produced them.
func TestAPIErrorIsUnauthorized(t *testing.T) {
	err := &APIError{StatusCode: http.StatusUnauthorized, Message: "bad key"}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatal("401 APIError must match ErrUnauthorized via errors.Is")
	}

	err = &APIError{StatusCode: http.StatusForbidden, Message: "no"}
	if errors.Is(err, ErrUnauthorized) {
		t.Fatal("non-401 APIError must not match ErrUnauthorized")
	}

	if !IsSentinelWrapped(ErrUnauthorized) {
		t.Fatal("ErrUnauthorized itself must match")
	}
}

// IsSentinelWrapped reports whether the sentinel matches itself (sanity check
// for errors.Is semantics).
func IsSentinelWrapped(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}
