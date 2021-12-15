package capact

import (
	"github.com/avast/retry-go"
)

// retryForFn retries failed function 5 times. The exponential retries are used: ~102ms ~302ms ~700ms 1.5s
func retryForFn(fn retry.RetryableFunc, customOpts ...retry.Option) error {
	opts := []retry.Option{
		retry.Attempts(10),
		retry.DelayType(retry.BackOffDelay),
	}
	opts = append(opts, customOpts...)

	return retry.Do(
		fn,
		opts...,
	)
}
