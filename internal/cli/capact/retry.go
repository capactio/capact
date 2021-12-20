package capact

import (
	"github.com/avast/retry-go"
)

// retryAttemptsCount defined number of retries of failed function.
const retryAttemptsCount = 10

// retryForFn retries failed function. The exponential retries are used: ~102ms ~302ms ~700ms 1.5s and so on.
func retryForFn(fn retry.RetryableFunc, customOpts ...retry.Option) error {
	opts := []retry.Option{
		retry.Attempts(retryAttemptsCount),
		retry.DelayType(retry.BackOffDelay),
	}
	opts = append(opts, customOpts...)

	return retry.Do(
		fn,
		opts...,
	)
}
