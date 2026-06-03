package client

import (
	"context"
	"tsunagi/src/rpc"
)

// SendEventRequest is a request to send the given event through an rpc stream.
// The provided resp chan can be used to confirm when the stream processes and sends this event.
// The provided ctx can be used for timeouts.
type SendEventRequest struct {

	// event is the rpc event to send. This must not be nil.
	event *rpc.Event

	// resp is a channel to report success/failures.
	// If nil it is ignored.
	resp chan error

	// ctx is the deadline for this request. If nil it is ignored.
	ctx context.Context
}

// PutErr sends this error message through the resp chan.
func (r *SendEventRequest) PutErr(err error) {
	if r.resp != nil {
		select {
		default:
			return
		case r.resp <- err:
			return
		}
	}
}

// IsDeadline checks if the context is done.
func (r *SendEventRequest) IsDeadline() bool {
	if r.ctx != nil {
		select {
		default:
			return false
		case <-r.ctx.Done():
			return true
		}
	}
	return false
}
