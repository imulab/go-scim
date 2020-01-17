package args

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// Retry returns a function that runs the specified callback as many times as it can (no more
// frequently than once every interval) until the timeout has past. The context passed to the
// callback is not canceled when the timeout expires.
func Retry(timeout time.Duration, interval time.Duration, callback func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var err error
		t := newInstantTicker(interval)
		defer t.Stop()
		timedCtx, timedCtxCancel := context.WithTimeout(ctx, timeout)
		for {
			select {
			case <-timedCtx.Done():
				timedCtxCancel()
				return errors.Wrap(err, "timeout exceded")
			case <-t.C:
				// Use the original context, not the one with the timeout for how long to retry for.
				err = callback(ctx)
				if err != nil {
					continue
				}
				timedCtxCancel()
				return nil
			}
		}
	}
}

func newInstantTicker(repeat time.Duration) *time.Ticker {
	ticker := time.NewTicker(repeat)
	oc := ticker.C
	nc := make(chan time.Time, 1)
	go func() {
		nc <- time.Now()
		for tm := range oc {
			nc <- tm
		}
	}()
	ticker.C = nc
	return ticker
}