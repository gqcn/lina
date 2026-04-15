// Package closeutil provides shared helpers for closing resources without
// dropping the returned error.
package closeutil

import "github.com/gogf/gf/v2/errors/gerror"

// Closer describes one resource that can be closed.
type Closer interface {
	Close() error
}

// Close folds one close error into errPtr when the caller already returns an error.
func Close(closer Closer, errPtr *error, action string) {
	if closer == nil {
		return
	}
	closeErr := closer.Close()
	if closeErr == nil {
		return
	}
	wrapped := gerror.Wrap(closeErr, action)
	if errPtr == nil {
		panic(wrapped)
	}
	if *errPtr == nil {
		*errPtr = wrapped
	}
}
