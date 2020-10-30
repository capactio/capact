package wait

import (
	"fmt"
	"os"
	"time"
)

const tickDuration = 1 * time.Second

// NoMoreThan waits for a function to finish until a given timeout is reached.
// When it receives stop signal, it exits without error.
func NoMoreThan(stopCh <-chan struct{}, fn func() error, timeout time.Duration, tickErrLogFn func(err error)) error {
	timeoutCh := time.After(timeout)
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	if tickErrLogFn == nil {
		// assign empty log function
		tickErrLogFn = func(err error) {}
	}

	var lastTickErr error
	for {
		select {
		case <-stopCh:
			os.Exit(0) // Nothing to do here, received SIGTERM/SIGINT signal
		case <-timeoutCh:
			return fmt.Errorf("timeout excedeed while waiting for success in given timeout %s: %w", timeout, lastTickErr)
		case <-ticker.C:
			err := fn()
			if err != nil {
				tickErrLogFn(err)
				lastTickErr = err
				break
			}

			return nil
		}
	}
}
