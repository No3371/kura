package advisor

import (
	"context"
	"errors"
)

// shopperError is a failed turn whose message is written for the shopper.
// Error keeps the cause for the server log; Public shows only msg.
type shopperError struct {
	msg   string
	cause error
}

func (e *shopperError) Error() string {
	if e.cause == nil {
		return e.msg
	}
	return e.msg + ": " + e.cause.Error()
}

func (e *shopperError) Unwrap() error { return e.cause }

func shopperErr(msg string, cause error) error { return &shopperError{msg, cause} }

// Public is the chat text for an error Chat returned. Anything not written
// for the shopper (SDK, network, decoding errors) gets a generic line: raw
// API detail and env var names are for the operator's log, not the panel.
func Public(err error) string {
	if e, ok := errors.AsType[*shopperError](err); ok {
		return e.msg
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "回覆時間過長，請把問題縮小一些再試一次"
	}
	return "顧問暫時無法回覆，請稍後再試"
}
