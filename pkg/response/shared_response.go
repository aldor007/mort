package response

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/aldor007/mort/pkg/transforms"
)

// SharedResponse wraps a Response with reference counting for safe sharing
// across multiple goroutines. This eliminates the need to create full copies
// of responses when distributing them to multiple consumers (e.g., in request collapsing).
type SharedResponse struct {
	resp     *Response
	refCount *atomic.Int32
	body     []byte      // Immutable shared buffer
	headers  http.Header // Immutable shared headers
}

// NewSharedResponse creates a shareable response from a buffered response.
// The response must be fully buffered (IsBuffered() == true) before creating a SharedResponse.
// Returns an error if the response is not buffered or if buffering fails.
func NewSharedResponse(resp *Response) (*SharedResponse, error) {
	if resp == nil {
		return nil, errors.New("response cannot be nil")
	}

	// Ensure response is buffered
	body, err := resp.Body()
	if err != nil {
		return nil, err
	}

	sr := &SharedResponse{
		resp:     resp,
		refCount: &atomic.Int32{},
		body:     body,
		headers:  resp.Headers.Clone(),
	}
	sr.refCount.Store(1)
	return sr, nil
}

// Acquire increments the reference count and returns a Response view
// that shares the underlying buffer. The returned Response is safe to use
// for reading but should not be modified.
//
// Each call to Acquire() must be matched with a corresponding Release() call.
func (sr *SharedResponse) Acquire() *Response {
	sr.refCount.Add(1)

	// Create lightweight Response view that shares the buffer
	view := &Response{
		StatusCode:    sr.resp.StatusCode,
		Headers:       sr.headers.Clone(), // Headers are small, clone for safety
		ContentLength: sr.resp.ContentLength,
		body:          sr.body, // Share the body buffer (read-only)
		debug:         sr.resp.debug,
		errorValue:    sr.resp.errorValue,
		cachable:      sr.resp.cachable,
		ttl:           sr.resp.ttl,
	}

	// Set up reader for the shared body
	view.bodySeeker = bytes.NewReader(sr.body)
	view.reader = io.NopCloser(view.bodySeeker)

	// Copy transforms if present
	if len(sr.resp.trans) > 0 {
		view.trans = make([]transforms.Transforms, len(sr.resp.trans))
		copy(view.trans, sr.resp.trans)
	}

	return view
}

// Release decrements the reference count.
// When the reference count reaches zero, the underlying resources are released.
//
// This method is safe to call from multiple goroutines and should always be
// called to prevent resource leaks. Use defer to ensure Release is called:
//
//	view := sharedResp.Acquire()
//	defer sharedResp.Release()
func (sr *SharedResponse) Release() {
	if sr == nil {
		return // Nil-safe for convenience
	}
	if sr.refCount.Add(-1) == 0 {
		// Last reference released, clean up resources
		if sr.resp != nil {
			sr.resp.Close()
		}
	}
}

// RefCount returns the current reference count.
// This method is primarily useful for testing and debugging.
func (sr *SharedResponse) RefCount() int32 {
	return sr.refCount.Load()
}
