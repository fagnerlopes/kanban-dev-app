package handler

import (
	"errors"
)

// panicToError converts a recovered panic value to an error for Sentry.
func panicToError(p any) error {
	if err, ok := p.(error); ok {
		return err
	}
	return errors.New("panic: " + stringify(p))
}

func stringify(p any) string {
	switch v := p.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return "unknown panic value"
	}
}
