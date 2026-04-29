package domain

import "errors"

// ErrNotFound is the canonical "entity does not exist" sentinel for the
// resume domain. Storage implementations return it instead of (nil, nil) so
// callers don't have to nil-check on top of err-check (the "nil-as-success"
// anti-pattern). Service layer maps it onto its own ErrNotFound when it
// needs to expose a service-level error contract.
var ErrNotFound = errors.New("entity not found")
