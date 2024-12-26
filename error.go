package raindrop

import "errors"

var (
	ErrConnectionNotFound = errors.New("connection not found")

	ErrConnectionClosed = errors.New("connection has been closed")

	ErrConnectionTimeout = errors.New("connection is timed out")
)
