package capact

import (
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

// retryForFn retries failed function 4 times. The exponential retries are used: ~102ms ~302ms ~700ms 1.5s
func retryForFn(fn retry.RetryableFunc) error {
	err := retry.Do(
		fn,
		retry.Attempts(5),
		retry.DelayType(retry.BackOffDelay),
	)
	if err != nil {
		return errors.Wrap(err, "while waiting")
	}

	return nil
}
