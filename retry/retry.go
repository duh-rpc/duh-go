package retry

import (
	"context"
	"time"
)

type Interval interface {
	Next() time.Duration
}

type BackOff struct {
	Min    time.Duration
	Max    time.Duration
	Factor float64
}

func (b BackOff) Next() time.Duration {
	// TODO Implement backoff reference implementations
	//  https://github.com/mailgun/holster/blob/master/retry/backoff.go
	//  https://github.com/kubernetes/client-go/blob/master/util/flowcontrol/backoff.go#L34
	return time.Second
}

var DefaultBackOff = BackOff{
	Min:    500 * time.Millisecond,
	Max:    5 * time.Second,
	Factor: 2,
}

type Sleep time.Duration

func (s Sleep) Next() time.Duration {
	return time.Duration(s)
}

type Policy struct {
	Interval interface{} // BackOff or Sleep
	OnCodes  []int64
	Attempts int // 0 for infinite
}

var Twice = Policy{
	Interval: DefaultBackOff,
	OnCodes:  []int64{409, 429, 500, 502, 503, 504},
	Attempts: 2,
}

var UntilSuccess = Policy{
	Interval: DefaultBackOff,
	OnCodes:  []int64{409, 429, 500, 502, 503, 504},
	Attempts: 0,
}

var OnConflict = Policy{
	Interval: DefaultBackOff,
	OnCodes:  []int64{409},
	Attempts: 0,
}

func shouldRetry(err error, policy Policy) bool {
	// You would implement this function based on how errors are structured in your application
	// to determine if a given error code should be retried.
	// For simplicity, I'm leaving it unimplemented.
	return false
}

func On(ctx context.Context, p Policy, operation func(context.Context, int) error) error {
	// TODO: Come back to this once the ABI is where I want it

	attempt := 0
	var sleepDuration time.Duration

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := operation(ctx, attempt)
			if err == nil || (p.Attempts != 0 && attempt >= p.Attempts) {
				return err
			}

			if shouldRetry(err, p) {
				switch interval := p.Interval.(type) {
				case BackOff:
					sleepDuration = time.Duration(float64(interval.Min) * float64(attempt))
					if sleepDuration > interval.Max {
						sleepDuration = interval.Max
					}
				case Sleep:
					sleepDuration = time.Duration(interval)
				default:
					return err
				}

				time.Sleep(sleepDuration)
				attempt++
			} else {
				return err
			}
		}
	}
}
