package locker

import "errors"

// Error definitions for the locker package.
var (
	// ErrLockNotHeld is returned when trying to renew a lock that is not held.
	ErrLockNotHeld = errors.New("lock not held by current node")

	// ErrRenewalFailed is returned when lease renewal fails.
	ErrRenewalFailed = errors.New("lease renewal failed")
)
