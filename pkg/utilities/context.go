package utilities

import (
	"context"
	"sync"
)

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

const (
	ContextKeyRequestID ContextKey = "request_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(ContextKeyRequestID).(string)
	if !ok {
		return ""
	}
	return requestID
}

// WaitForCleanup waits for the provided sync.WaitGroup to be empty, or for the
// provided context to be canceled. This is useful for waiting for a set of
// asynchronous operations to complete before proceeding.
func WaitForCleanup(ctx context.Context, wg *sync.WaitGroup) {
	doneChannel := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChannel)
	}()
	select {
	case <-ctx.Done():
		return
	case <-doneChannel:
		return
	}
}
